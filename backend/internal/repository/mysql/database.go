package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	mysqlprotocol "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Open creates the application's GORM connection. It never changes the schema.
func Open(dsn string) (*gorm.DB, *sql.DB, error) {
	normalizedDSN, err := normalizeDSN(dsn)
	if err != nil {
		return nil, nil, err
	}
	database, err := gorm.Open(gormmysql.Open(normalizedDSN), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger: gormlogger.New(log.New(os.Stderr, "", log.LstdFlags), gormlogger.Config{
			LogLevel: gormlogger.Error, IgnoreRecordNotFoundError: true, ParameterizedQueries: true, Colorful: false,
		}),
		DisableAutomaticPing: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open MySQL: %w", err)
	}

	sqlDatabase, err := database.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("obtain MySQL pool: %w", err)
	}
	sqlDatabase.SetMaxOpenConns(10)
	sqlDatabase.SetMaxIdleConns(5)
	sqlDatabase.SetConnMaxLifetime(5 * time.Minute)
	sqlDatabase.SetConnMaxIdleTime(time.Minute)

	return database, sqlDatabase, nil
}

func normalizeDSN(raw string) (string, error) {
	configuration, err := mysqlprotocol.ParseDSN(raw)
	if err != nil {
		return "", fmt.Errorf("parse MySQL DSN: %w", err)
	}
	configuration.ParseTime = true
	configuration.Loc = time.UTC
	configuration.Timeout = 5 * time.Second
	configuration.ReadTimeout = 10 * time.Second
	configuration.WriteTimeout = 10 * time.Second
	return configuration.FormatDSN(), nil
}
