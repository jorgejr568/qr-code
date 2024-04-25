package cfg

import (
	goenv "github.com/Netflix/go-env"
	"go.uber.org/zap"
)

type Config struct {
	Port int `env:"PORT,default=8080"`

	DefaultQrCodeSize int `env:"DEFAULT_QR_CODE_SIZE,default=256"`
	MaxQrCodeSize     int `env:"MAX_QR_CODE_SIZE,default=2048"`
	MinQrCodeSize     int `env:"MIN_QR_CODE_SIZE,default=64"`

	LogLevel     string `env:"LOG_LEVEL,default=debug"`
	PProfEnabled bool   `env:"PPROF_ENABLED,default=false"`
}

func NewConfig() (*Config, error) {
	var c Config
	_, err := goenv.UnmarshalFromEnviron(&c)
	if err != nil {
		zap.L().Error("failed to load environment variables", zap.Error(err))
		return nil, err
	}

	return &c, nil
}
