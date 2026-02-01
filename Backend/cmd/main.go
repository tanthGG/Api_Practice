package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"go-api-practice/config"
	"go-api-practice/internal/handlers"
	"go-api-practice/internal/repositories"
	"go-api-practice/internal/services"
)

func main() {
	// Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("failed to load config: %v", err)
	}
	// Logger
	logger := setupLogger(cfg.Log.Level, cfg.Log.Env)

	// DB
	db, err := setupDatabase(cfg.DbSQL)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Echo
	server := echo.New()
	server.HideBanner = false
	server.Use(middleware.Logger())
	server.Use(middleware.Recover())
	server.Use(middleware.CORS())

	healthHandler := handlers.NewHealthCheckHandler()
	loanRepo := repositories.NewLoanRepository(db)
	loanService := services.NewLoanService(loanRepo, logger)

	httpServer := handlers.NewHttpServer(cfg, server, healthHandler, logger, loanService)

	// Start
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	go func() {
		logger.Infof("skill API listening on %s", addr)
		if err := httpServer.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	// Grateful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Info("shutdown signal received, closing server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Server().Shutdown(shutdownCtx); err != nil {
		logger.Errorf("graceful shutdown failed: %v", err)
	}
	logger.Info("server exited")
}

func setupDatabase(cfg config.DbSQLConfig) (*sql.DB, error) {
	dsn := cfg.GetDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxLifeTimeMinutes > 0 {
		db.SetConnMaxLifetime(cfg.MaxLifeTimeMinutes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func setupLogger(level, env string) *logrus.Logger {
	logger := logrus.New()
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
	if env == "prod" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}
	return logger
}
