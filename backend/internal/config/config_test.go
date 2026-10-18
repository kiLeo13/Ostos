package config

import (
	"reflect"
	"testing"
)

func TestLoadRequiresValidSnowflakeNodeAndDatabase(t *testing.T) {
	tests := []map[string]string{
		{"OSTOS_SNOWFLAKE_NODE_ID": "0"},
		{"OSTOS_DATABASE_DSN": "dsn"},
		{"OSTOS_DATABASE_DSN": "dsn", "OSTOS_SNOWFLAKE_NODE_ID": "-1"},
		{"OSTOS_DATABASE_DSN": "dsn", "OSTOS_SNOWFLAKE_NODE_ID": "1024"},
		{"OSTOS_DATABASE_DSN": "dsn", "OSTOS_SNOWFLAKE_NODE_ID": "not-a-number"},
	}
	for index, values := range tests {
		_, err := Load(func(key string) (string, bool) { value, ok := values[key]; return value, ok })
		if err == nil {
			t.Fatalf("case %d unexpectedly succeeded", index)
		}
	}
}

func TestLoadRequiresProductionTLSAndTrustedProxies(t *testing.T) {
	base := map[string]string{
		"OSTOS_ENV": "production", "OSTOS_DATABASE_DSN": "dsn", "OSTOS_SNOWFLAKE_NODE_ID": "1",
		"OSTOS_JWT_SECRET_BASE64":      "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=",
		"OSTOS_EMAIL_TOKEN_KEY_BASE64": "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=",
	}
	for _, missing := range []string{"OSTOS_TLS_CERT_FILE", "OSTOS_TLS_KEY_FILE", "OSTOS_API_HOSTNAME", "OSTOS_TRUSTED_PROXY_CIDRS"} {
		values := cloneMap(base)
		values["OSTOS_TLS_CERT_FILE"] = "/cert"
		values["OSTOS_TLS_KEY_FILE"] = "/key"
		values["OSTOS_API_HOSTNAME"] = "api.example.com"
		values["OSTOS_TRUSTED_PROXY_CIDRS"] = "203.0.113.0/24"
		delete(values, missing)
		if _, err := Load(func(key string) (string, bool) { value, ok := values[key]; return value, ok }); err == nil {
			t.Fatalf("missing %s unexpectedly succeeded", missing)
		}
	}
}

func TestLoadParsesExactOriginsAndTrustedProxyCIDRs(t *testing.T) {
	values := map[string]string{
		"OSTOS_DATABASE_DSN": "dsn", "OSTOS_SNOWFLAKE_NODE_ID": "1",
		"OSTOS_CORS_ORIGINS":        "https://app.example.com,http://localhost:5173",
		"OSTOS_TRUSTED_PROXY_CIDRS": "203.0.113.7/24, 2001:db8::/32",
	}
	loaded, err := Load(func(key string) (string, bool) { value, ok := values[key]; return value, ok })
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !reflect.DeepEqual(loaded.CORSOrigins, []string{"https://app.example.com", "http://localhost:5173"}) {
		t.Fatalf("origins = %#v", loaded.CORSOrigins)
	}
	if got := loaded.TrustedProxies[0].String(); got != "203.0.113.0/24" {
		t.Fatalf("masked proxy prefix = %s", got)
	}
}

func TestLoadRejectsInvalidOriginsAndProxyCIDRs(t *testing.T) {
	for key, value := range map[string]string{
		"OSTOS_CORS_ORIGINS":        "https://example.com/path",
		"OSTOS_TRUSTED_PROXY_CIDRS": "not-a-cidr",
	} {
		values := map[string]string{"OSTOS_DATABASE_DSN": "dsn", "OSTOS_SNOWFLAKE_NODE_ID": "1", key: value}
		if _, err := Load(func(key string) (string, bool) { value, ok := values[key]; return value, ok }); err == nil {
			t.Fatalf("invalid %s unexpectedly succeeded", key)
		}
	}
}

func cloneMap(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func TestLoadUsesSafeDefaults(t *testing.T) {
	values := map[string]string{"OSTOS_DATABASE_DSN": "dsn", "OSTOS_SNOWFLAKE_NODE_ID": "1023"}
	loaded, err := Load(func(key string) (string, bool) { value, ok := values[key]; return value, ok })
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.SnowflakeNode != 1023 || loaded.HTTPAddress != ":8080" || loaded.BodyLimit != 1<<20 {
		t.Fatalf("unexpected configuration: %+v", loaded)
	}
}
