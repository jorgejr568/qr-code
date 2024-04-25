package main

import (
	"context"
	"fmt"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"qr-code-server/cfg"
	httpserver "qr-code-server/internal/http"
	usecases "qr-code-server/internal/use-cases"
)

func main() {
	ctx := context.Background()
	container := dig.New()
	logger := zap.Must(zap.NewProduction()).With(zap.String("service", "qr-code-server"))
	defer logger.Sync()

	container.Provide(func() *zap.Logger {
		return logger
	})

	err := container.Provide(cfg.NewConfig)
	if err != nil {
		logger.Error("failed to provide config", zap.Error(err))
		return
	}

	err = usecases.Provide(container)
	if err != nil {
		logger.Error("failed to provide use cases", zap.Error(err))
		return
	}

	err = httpserver.Provide(container)
	if err != nil {
		logger.Error("failed to provide http server", zap.Error(err))
		return
	}

	err = container.Invoke(func(httpServer httpserver.Server, c *cfg.Config, logger *zap.Logger) {
		logger.Info("starting server", zap.Int("port", c.Port))
		err := httpServer.ListenAndServeWithGracefulShutdown(ctx, fmt.Sprintf(":%d", c.Port))
		if err != nil {
			logger.Error("failed to start server", zap.Error(err))
		}

		logger.Info("server stopped")
	})
}
