package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ostos/backend/internal/config"
	"ostos/backend/internal/platform"
	mysqlrepository "ostos/backend/internal/repository/mysql"
	"ostos/backend/internal/transport/http/router"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	configuration, err := config.Load(os.LookupEnv)
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	if _, err := platform.NewSnowflakeGenerator(configuration.SnowflakeNode); err != nil {
		logger.Error("initialize Snowflake generator", "error", err)
		os.Exit(1)
	}

	_, sqlDatabase, err := mysqlrepository.Open(configuration.DatabaseDSN)
	if err != nil {
		logger.Error("initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := sqlDatabase.Close(); closeErr != nil {
			logger.Error("close database", "error", closeErr)
		}
	}()

	server := &http.Server{
		Addr:              configuration.HTTPAddress,
		Handler:           router.New(router.Dependencies{Logger: logger, Database: sqlDatabase, BodyLimit: configuration.BodyLimit}),
		ReadHeaderTimeout: configuration.ReadTimeout,
		ReadTimeout:       configuration.ReadTimeout,
		WriteTimeout:      configuration.WriteTimeout,
		IdleTimeout:       60 * time.Second,
	}

	shutdownContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		logger.Info("API listening", "address", configuration.HTTPAddress, "environment", configuration.Environment)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Error("API server stopped unexpectedly", "error", serveErr)
			stop()
		}
	}()

	<-shutdownContext.Done()
	gracefulContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(gracefulContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}
