package mysql_test

import (
	"context"
	"os"
	"testing"
	"time"

	mysqlrepository "ostos/backend/internal/repository/mysql"
)

func TestOpenConnectsToMigratedRealMySQLWithoutSchemaMutation(t *testing.T) {
	dsn := os.Getenv("OSTOS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OSTOS_TEST_MYSQL_DSN is not set; run the real-MySQL Compose integration test")
	}

	_, database, err := mysqlrepository.Open(dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}

	var schemaVersionTableCount int
	if err := database.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name = 'schema_migrations'
	`).Scan(&schemaVersionTableCount); err != nil {
		t.Fatalf("inspect schema: %v", err)
	}
	if schemaVersionTableCount != 1 {
		t.Fatalf("schema_migrations table count = %d", schemaVersionTableCount)
	}

	var capabilityTableCount int
	if err := database.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name <> 'schema_migrations'
	`).Scan(&capabilityTableCount); err != nil {
		t.Fatalf("inspect capability tables: %v", err)
	}
	if capabilityTableCount != 0 {
		t.Fatalf("foundation unexpectedly created %d capability tables", capabilityTableCount)
	}
}
