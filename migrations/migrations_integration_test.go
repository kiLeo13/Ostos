package migrations_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"ostos/migrations"
)

func TestApplyMigratesEmptyMySQLAndIsIdempotent(t *testing.T) {
	dsn := os.Getenv("OSTOS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OSTOS_TEST_MYSQL_DSN is not set; run the real-MySQL Compose integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := migrations.Apply(ctx, dsn); err != nil {
		t.Fatalf("first migration: %v", err)
	}
	if err := migrations.Apply(ctx, dsn); err != nil {
		t.Fatalf("idempotent migration: %v", err)
	}

	database, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open migrated database: %v", err)
	}
	defer database.Close()
	var version uint
	var dirty bool
	if err := database.QueryRowContext(ctx, "SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != 1 || dirty {
		t.Fatalf("unexpected schema state: version=%d dirty=%t", version, dirty)
	}
}
