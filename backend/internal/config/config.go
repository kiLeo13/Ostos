package config

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

const defaultBodyLimit int64 = 1 << 20

// Config contains API process configuration. Secrets are never logged by this package.
type Config struct {
	Environment   string
	HTTPAddress   string
	DatabaseDSN   string
	SnowflakeNode int64
	BodyLimit     int64
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
}

type LookupFunc func(string) (string, bool)

func Load(lookup LookupFunc) (Config, error) {
	if lookup == nil {
		return Config{}, errors.New("configuration lookup is required")
	}

	config := Config{
		Environment:  valueOrDefault(lookup, "OSTOS_ENV", "development"),
		HTTPAddress:  valueOrDefault(lookup, "OSTOS_HTTP_ADDRESS", ":8080"),
		BodyLimit:    defaultBodyLimit,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	databaseDSN, ok := lookup("OSTOS_DATABASE_DSN")
	if !ok || databaseDSN == "" {
		return Config{}, errors.New("OSTOS_DATABASE_DSN is required")
	}
	config.DatabaseDSN = databaseDSN

	nodeValue, ok := lookup("OSTOS_SNOWFLAKE_NODE_ID")
	if !ok || nodeValue == "" {
		return Config{}, errors.New("OSTOS_SNOWFLAKE_NODE_ID is required")
	}
	nodeID, err := strconv.ParseInt(nodeValue, 10, 64)
	if err != nil || nodeID < 0 || nodeID > 1023 {
		return Config{}, fmt.Errorf("OSTOS_SNOWFLAKE_NODE_ID must be an integer from 0 through 1023")
	}
	config.SnowflakeNode = nodeID

	return config, nil
}

func valueOrDefault(lookup LookupFunc, key, fallback string) string {
	if value, ok := lookup(key); ok && value != "" {
		return value
	}
	return fallback
}
