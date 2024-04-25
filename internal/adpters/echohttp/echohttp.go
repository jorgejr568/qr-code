package echohttp

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func HttpHandlerToEchoHandler(httpHandler func(http.ResponseWriter, *http.Request)) echo.HandlerFunc {
	return func(c echo.Context) error {
		httpHandler(c.Response().Writer, c.Request())
		return nil
	}
}
