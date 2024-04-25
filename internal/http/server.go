package http

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"qr-code-server/cfg"
	"qr-code-server/internal/adpters/echohttp"
	"qr-code-server/internal/adpters/echozap"
	usecases "qr-code-server/internal/use-cases"
	"strconv"
)

type Server interface {
	ListenAndServeWithGracefulShutdown(ctx context.Context, addr string) error
}

type echoServer struct {
	e                     *echo.Echo
	logger                *zap.Logger
	pprofEnabled          bool
	generateQrCodeUseCase usecases.GenerateQrCodeFromData
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
	if s.pprofEnabled {
		s.logger.Debug("pprof is enabled")
		s.registerPprof()
	}
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

func (s echoServer) registerPprof() {
	s.e.GET("/debug/pprof", echohttp.HttpHandlerToEchoHandler(pprof.Index))
	s.e.GET("/debug/cmdline", echohttp.HttpHandlerToEchoHandler(pprof.Cmdline))
	s.e.GET("/debug/profile", echohttp.HttpHandlerToEchoHandler(pprof.Profile))
	s.e.GET("/debug/symbol", echohttp.HttpHandlerToEchoHandler(pprof.Symbol))
	s.e.GET("/debug/trace", echohttp.HttpHandlerToEchoHandler(pprof.Trace))
	s.e.GET("/debug/allocs", echohttp.HttpHandlerToEchoHandler(pprof.Handler("allocs").ServeHTTP))
	s.e.GET("/debug/block", echohttp.HttpHandlerToEchoHandler(pprof.Handler("block").ServeHTTP))
	s.e.GET("/debug/goroutine", echohttp.HttpHandlerToEchoHandler(pprof.Handler("goroutine").ServeHTTP))
	s.e.GET("/debug/heap", echohttp.HttpHandlerToEchoHandler(pprof.Handler("heap").ServeHTTP))
	s.e.GET("/debug/mutex", echohttp.HttpHandlerToEchoHandler(pprof.Handler("mutex").ServeHTTP))
	s.e.GET("/debug/threadcreate", echohttp.HttpHandlerToEchoHandler(pprof.Handler("threadcreate").ServeHTTP))
}

func newEchoServer(c *cfg.Config, logger *zap.Logger, generateQrCodeFromUrl usecases.GenerateQrCodeFromData) Server {
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
		pprofEnabled:          c.PProfEnabled,
	}
}
