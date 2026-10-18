package config

import (
	"encoding/base64"
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
	JWTSecret     []byte
	JWTIssuer     string
	JWTAudience   string
	EmailTokenKey []byte
	PublicAppURL  string
	SESEnabled    bool
	SESRegion     string
	SESFrom       string
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
		JWTIssuer:    valueOrDefault(lookup, "OSTOS_JWT_ISSUER", "ostos"),
		JWTAudience:  valueOrDefault(lookup, "OSTOS_JWT_AUDIENCE", "ostos-clients"),
		PublicAppURL: valueOrDefault(lookup, "OSTOS_PUBLIC_APP_URL", "http://localhost:8081"),
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

	production := config.Environment == "production"
	jwtDefault := base64.StdEncoding.EncodeToString([]byte("local-only-jwt-secret-change-me!"))
	emailDefault := base64.StdEncoding.EncodeToString([]byte("local-only-email-key-change-me!!"))
	jwtEncoded := valueOrDefault(lookup, "OSTOS_JWT_SECRET_BASE64", jwtDefault)
	emailEncoded := valueOrDefault(lookup, "OSTOS_EMAIL_TOKEN_KEY_BASE64", emailDefault)
	if production {
		if _, ok := lookup("OSTOS_JWT_SECRET_BASE64"); !ok {
			return Config{}, errors.New("OSTOS_JWT_SECRET_BASE64 is required in production")
		}
		if _, ok := lookup("OSTOS_EMAIL_TOKEN_KEY_BASE64"); !ok {
			return Config{}, errors.New("OSTOS_EMAIL_TOKEN_KEY_BASE64 is required in production")
		}
	}
	config.JWTSecret, err = decodeSecret(jwtEncoded, 32, "OSTOS_JWT_SECRET_BASE64")
	if err != nil {
		return Config{}, err
	}
	config.EmailTokenKey, err = decodeSecret(emailEncoded, 32, "OSTOS_EMAIL_TOKEN_KEY_BASE64")
	if err != nil {
		return Config{}, err
	}
	config.SESEnabled, err = strconv.ParseBool(valueOrDefault(lookup, "OSTOS_SES_ENABLED", "false"))
	if err != nil {
		return Config{}, errors.New("OSTOS_SES_ENABLED must be true or false")
	}
	config.SESRegion = valueOrDefault(lookup, "OSTOS_SES_REGION", "us-east-1")
	if from, ok := lookup("OSTOS_SES_FROM"); ok {
		config.SESFrom = from
	}
	if config.SESEnabled && config.SESFrom == "" {
		return Config{}, errors.New("OSTOS_SES_FROM is required when SES delivery is enabled")
	}

	return config, nil
}

func decodeSecret(encoded string, minimum int, name string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(decoded) < minimum {
		return nil, fmt.Errorf("%s must be standard base64 containing at least %d bytes", name, minimum)
	}
	return decoded, nil
}

func valueOrDefault(lookup LookupFunc, key, fallback string) string {
	if value, ok := lookup(key); ok && value != "" {
		return value
	}
	return fallback
}
