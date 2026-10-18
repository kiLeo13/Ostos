package migrations

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/*.up.sql
var files embed.FS

// Apply migrates the configured MySQL schema forward to the latest version.
func Apply(ctx context.Context, rawDSN string) error {
	database, err := openDatabase(ctx, rawDSN)
	if err != nil {
		return err
	}
	defer database.Close()

	driver, err := migratemysql.WithInstance(database, &migratemysql.Config{})
	if err != nil {
		return fmt.Errorf("create migration database driver: %w", err)
	}
	source, err := iofs.New(files, "sql")
	if err != nil {
		return fmt.Errorf("create embedded migration source: %w", err)
	}
	runner, err := migrate.NewWithInstance("iofs", source, "mysql", driver)
	if err != nil {
		return fmt.Errorf("create migration runner: %w", err)
	}

	if err := runner.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

func openDatabase(ctx context.Context, rawDSN string) (*sql.DB, error) {
	configuration, err := mysqldriver.ParseDSN(rawDSN)
	if err != nil {
		return nil, fmt.Errorf("parse migration database DSN: %w", err)
	}
	configuration.ParseTime = true
	configuration.Loc = time.UTC
	configuration.Timeout = 5 * time.Second
	configuration.ReadTimeout = 10 * time.Second
	configuration.WriteTimeout = 10 * time.Second
	configuration.MultiStatements = true

	database, err := sql.Open("mysql", configuration.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open migration database: %w", err)
	}
	database.SetMaxOpenConns(2)
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(time.Minute)

	pingContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := database.PingContext(pingContext); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping migration database: %w", err)
	}
	return database, nil
}
