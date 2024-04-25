package use_cases

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
	"io"
	"qr-code-server/cfg"
	"qr-code-server/internal/utils/validators"
)

type GenerateQrCodeFromUrl interface {
	Make(ctx context.Context, url string, size int) (io.Reader, error)
}

type skip2GenerateQrCodeFromUrl struct {
	defaultQrCodeSize int
	maxQrCodeSize     int
	minQrCodeSize     int
	logger            *zap.Logger
}

var (
	ErrorSizeMustBeBetweenMinAndMax = errors.New("invalid_size")
	ErrorUrlIsEmptyOrInvalid        = errors.New("invalid_url")
)

func (s skip2GenerateQrCodeFromUrl) Make(ctx context.Context, url string, size int) (io.Reader, error) {
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

	if url == "" || !validators.ValidateUrl(url) {
		s.logger.Warn("url is empty or invalid", zap.String("url", url))
		return nil, ErrorUrlIsEmptyOrInvalid
	}

	imageBytes, err := qrcode.Encode(url, qrcode.Highest, size)

	if err != nil {
		s.logger.Error("failed to generate qr code", zap.Error(err))
		return nil, errors.Wrap(err, "failed to generate qr code")
	}

	imageBuffer := bytes.NewBuffer(imageBytes)
	return imageBuffer, nil
}

func newGenerateQrCodeFromUrl(c *cfg.Config, logger *zap.Logger) GenerateQrCodeFromUrl {
	return &skip2GenerateQrCodeFromUrl{
		defaultQrCodeSize: c.DefaultQrCodeSize,
		maxQrCodeSize:     c.MaxQrCodeSize,
		minQrCodeSize:     c.MinQrCodeSize,
		logger:            logger.With(zap.String("use-case", "generate-qr-code-from-url")),
	}
}
