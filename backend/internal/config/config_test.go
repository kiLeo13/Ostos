package config

import "testing"

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
