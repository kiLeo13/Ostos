package mysql

import (
	"testing"
	"time"

	mysqlprotocol "github.com/go-sql-driver/mysql"
)

func TestNormalizeDSNEnforcesExactTimeAndBoundedNetworkOperations(t *testing.T) {
	normalized, err := normalizeDSN("user:password@tcp(database:3306)/ostos?tls=true")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	configuration, err := mysqlprotocol.ParseDSN(normalized)
	if err != nil {
		t.Fatalf("parse normalized DSN: %v", err)
	}
	if !configuration.ParseTime || configuration.Loc != time.UTC {
		t.Fatalf("time configuration: parseTime=%t location=%v", configuration.ParseTime, configuration.Loc)
	}
	if configuration.Timeout != 5*time.Second || configuration.ReadTimeout != 10*time.Second || configuration.WriteTimeout != 10*time.Second {
		t.Fatalf("timeouts: connect=%s read=%s write=%s", configuration.Timeout, configuration.ReadTimeout, configuration.WriteTimeout)
	}
	if configuration.TLSConfig != "true" {
		t.Fatalf("TLS configuration was not preserved: %q", configuration.TLSConfig)
	}
}
