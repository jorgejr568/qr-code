package http

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"qr-code-server/internal/adpters/echozap"
	usecases "qr-code-server/internal/use-cases"
	"strconv"
)

type Server interface {
	ListenAndServeWithGracefulShutdown(ctx context.Context, addr string) error
}

type echoServer struct {
	e                     *echo.Echo
	generateQrCodeUseCase usecases.GenerateQrCodeFromData
	logger                *zap.Logger
}

func (s echoServer) ListenAndServeWithGracefulShutdown(ctx context.Context, addr string) error {
	interruptChain := make(chan os.Signal, 1)
	signal.Notify(interruptChain, os.Interrupt, os.Kill)
	go func() {
		<-interruptChain
		s.logger.Info("shutting down server")

		err := s.e.Shutdown(ctx)
		if err != nil {
			s.logger.Error("failed to shutdown server", zap.Error(err))
			return
		}

		s.logger.Info("server shutdown gracefully")
	}()

	s.register()
	err := s.e.Start(addr)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "failed to start server")
	}
	return nil
}

func (s echoServer) register() {
	s.e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "ok")
	})

	s.e.GET("/", func(c echo.Context) error {
		size := c.QueryParam("size")
		data := c.QueryParam("data")

		sizeInt := 0
		if size != "" {
			var err error
			sizeInt, err = strconv.Atoi(size)
			if err != nil {
				return c.JSON(http.StatusBadRequest, "size must be an integer")
			}
		}

		image, err := s.generateQrCodeUseCase.Make(c.Request().Context(), data, sizeInt)
		if err != nil {
			if errors.Is(err, usecases.ErrorSizeMustBeBetweenMinAndMax) || errors.Is(err, usecases.ErrorDataIsEmpty) {
				return c.JSON(http.StatusBadRequest, err.Error())
			}

			return c.JSON(http.StatusInternalServerError, "failed to generate qr code")
		}

		return c.Stream(http.StatusOK, "image/png", image)
	})
}

func newEchoServer(logger *zap.Logger, generateQrCodeFromUrl usecases.GenerateQrCodeFromData) Server {
	e := echo.New()
	e.Use(echozap.ZapLogger(logger))
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	e.HideBanner = true
	e.HidePort = true

	return &echoServer{
		e:                     e,
		generateQrCodeUseCase: generateQrCodeFromUrl,
		logger:                logger,
	}
}
