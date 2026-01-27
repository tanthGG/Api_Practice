package handlers

import (
	"github.com/labstack/echo/v4"
)

type IHealthCheckHandler interface {
	HealthCheck(c echo.Context) error
	ReadinessCheck(c echo.Context) error
}

type HealthCheckHandler struct{}

func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

func (h *HealthCheckHandler) HealthCheck(c echo.Context) error {
	return c.String(200, "ok")
}

func (h *HealthCheckHandler) ReadinessCheck(c echo.Context) error {
	return c.String(200, "ok")
}
