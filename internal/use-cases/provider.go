package use_cases

import (
	"go.uber.org/dig"
)

func Provide(container *dig.Container) error {
	err := container.Provide(newGenerateQrCodeFromUrl)
	if err != nil {
		return err
	}

	return nil
}
