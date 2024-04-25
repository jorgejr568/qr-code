package http

import (
	"go.uber.org/dig"
)

func Provide(container *dig.Container) error {
	err := container.Provide(newEchoServer)
	if err != nil {
		return err
	}

	return nil
}
