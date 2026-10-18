package platform

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimestampNormalizesToUTCAndRFC3339(t *testing.T) {
	location := time.FixedZone("test", -3*60*60)
	timestamp := NewTimestamp(time.Date(2026, 10, 17, 9, 0, 0, 123000000, location))
	encoded, err := json.Marshal(timestamp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(encoded) != `"2026-10-17T12:00:00.123Z"` {
		t.Fatalf("encoded = %s", encoded)
	}

	var decoded Timestamp
	if err := json.Unmarshal([]byte(`"2026-10-17T09:00:00-03:00"`), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Location() != time.UTC || decoded.Hour() != 12 {
		t.Fatalf("decoded timestamp was not normalized: %s", decoded.Time)
	}
}
