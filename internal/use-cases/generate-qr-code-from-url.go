package use_cases

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
	"image/png"
	"io"
	"qr-code-server/cfg"
)

type GenerateQrCodeFromData interface {
	Make(ctx context.Context, url string, size int) (io.Reader, error)
}

type skip2GenerateQrCodeFromData struct {
	defaultQrCodeSize int
	maxQrCodeSize     int
	minQrCodeSize     int
	logger            *zap.Logger
}

var (
	ErrorSizeMustBeBetweenMinAndMax = errors.New("invalid_size")
	ErrorDataIsEmpty                = errors.New("invalid_data")
)

func (s skip2GenerateQrCodeFromData) Make(ctx context.Context, data string, size int) (io.Reader, error) {
	if size == 0 {
		size = s.defaultQrCodeSize
	}

	if size < s.minQrCodeSize || size > s.maxQrCodeSize {
		s.
			logger.
			Warn(
				"size must be between min and max",
				zap.Int("size", size), zap.Int("min", s.minQrCodeSize), zap.Int("max", s.maxQrCodeSize),
			)
		return nil, fmt.Errorf("%w: min=%d, max=%d", ErrorSizeMustBeBetweenMinAndMax, s.minQrCodeSize, s.maxQrCodeSize)
	}

	if data == "" {
		s.logger.Warn("data is empty")
		return nil, ErrorDataIsEmpty
	}

	buffer, err := s.generateQrCodeFromData(ctx, data, size)

	if err != nil {
		s.logger.Error("failed to generate qr code", zap.Error(err))
		return nil, errors.Wrap(err, "failed to generate qr code")
	}

	return buffer, nil
}

func (s skip2GenerateQrCodeFromData) generateQrCodeFromData(ctx context.Context, url string, size int) (io.Reader, error) {
	qr, err := qrcode.New(url, qrcode.Highest)
	if err != nil {
		return nil, err
	}

	img := qr.Image(size)
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	var buffer bytes.Buffer
	err = encoder.Encode(&buffer, img)

	if err != nil {
		return nil, err
	}

	return &buffer, nil
}

func newGenerateQrCodeFromUrl(c *cfg.Config, logger *zap.Logger) GenerateQrCodeFromData {
	return &skip2GenerateQrCodeFromData{
		defaultQrCodeSize: c.DefaultQrCodeSize,
		maxQrCodeSize:     c.MaxQrCodeSize,
		minQrCodeSize:     c.MinQrCodeSize,
		logger:            logger.With(zap.String("use-case", "generate-qr-code-from-url")),
	}
}
