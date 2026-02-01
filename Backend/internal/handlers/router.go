package handlers

import (
	"go-api-practice/config"
	"go-api-practice/internal/handlers/rest"
	"go-api-practice/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type HttpServer struct {
	config             *config.Config
	server             *echo.Echo
	healthCheckHandler IHealthCheckHandler
	logger             *logrus.Logger
	loanService        services.LoanService
}

func NewHttpServer(
	config *config.Config,
	server *echo.Echo,
	healthCheckHandler IHealthCheckHandler,
	logger *logrus.Logger,
	loanService services.LoanService,
) *HttpServer {
	httpServer := &HttpServer{
		config:             config,
		server:             server,
		healthCheckHandler: healthCheckHandler,
		logger:             logger,
		loanService:        loanService,
	}
	httpServer.initRoute()

	return httpServer
}

func (s *HttpServer) initRoute() {
	e := s.server

	e.GET("/health", s.healthCheckHandler.HealthCheck)

	api := e.Group("/api")
	v1 := api.Group("/v1")
	loanHandler := rest.NewLoanHandler(s.logger, s.loanService)
	loanGroup := v1.Group("/loans")
	loanGroup.POST("", loanHandler.Apply)

}

func (s *HttpServer) Start(address string) error {
	return s.server.Start(address)
}

func (s *HttpServer) Server() *echo.Echo {
	return s.server
}
