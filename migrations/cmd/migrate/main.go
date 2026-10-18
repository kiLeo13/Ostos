package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ostos/migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	dsn, ok := os.LookupEnv("OSTOS_MIGRATION_DATABASE_DSN")
	if !ok || dsn == "" {
		logger.Error("OSTOS_MIGRATION_DATABASE_DSN is required")
		os.Exit(1)
	}

	signalContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(signalContext, 2*time.Minute)
	defer cancel()
	if err := migrations.Apply(ctx, dsn); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}
	logger.Info("database is at the latest schema version")
}
