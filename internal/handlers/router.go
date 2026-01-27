package handlers

import (
	"go-api-practice/config"
	skillshandlers "go-api-practice/internal/handlers/skills"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type HttpServer struct {
	config             *config.Config
	server             *echo.Echo
	healthCheckHandler IHealthCheckHandler
	skillHandler       *skillshandlers.SkillHandler
	logger             *logrus.Logger
}

func NewHttpServer(
	config *config.Config,
	server *echo.Echo,
	healthCheckHandler IHealthCheckHandler,
	skillHandler *skillshandlers.SkillHandler,
	logger *logrus.Logger,
) *HttpServer {
	httpServer := &HttpServer{
		config:             config,
		server:             server,
		healthCheckHandler: healthCheckHandler,
		skillHandler:       skillHandler,
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
	skills := v1.Group("/skills")
	skills.POST("", s.skillHandler.CreateSkill)

}

func (s *HttpServer) Start(address string) error {
	return s.server.Start(address)
}

func (s *HttpServer) Server() *echo.Echo {
	return s.server
}
