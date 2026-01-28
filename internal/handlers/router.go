package handlers

import (
	"go-api-practice/config"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type HttpServer struct {
	config             *config.Config
	server             *echo.Echo
	healthCheckHandler IHealthCheckHandler
	logger             *logrus.Logger
}

func NewHttpServer(
	config *config.Config,
	server *echo.Echo,
	healthCheckHandler IHealthCheckHandler,
	logger *logrus.Logger,
) *HttpServer {
	httpServer := &HttpServer{
		config:             config,
		server:             server,
		healthCheckHandler: healthCheckHandler,
		logger:             logger,
	}
	httpServer.initRoute()

	return httpServer
}

func (s *HttpServer) initRoute() {
	e := s.server

	e.GET("/health", s.healthCheckHandler.HealthCheck)

	api := e.Group("/api")
	v1 := api.Group("/v1")
	_ = v1.Group("/loan")

}

func (s *HttpServer) Start(address string) error {
	return s.server.Start(address)
}

func (s *HttpServer) Server() *echo.Echo {
	return s.server
}
