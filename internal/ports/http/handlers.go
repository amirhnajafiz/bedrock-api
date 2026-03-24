package http

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h HTTPServer) health(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h HTTPServer) createSession(c *echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not implemented")
}

func (h HTTPServer) updateSession(c *echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not implemented")
}

func (h HTTPServer) getSessions(c *echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not implemented")
}

func (h HTTPServer) getSessionLogs(c *echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not implemented")
}
