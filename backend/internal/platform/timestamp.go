package platform

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Timestamp normalizes application and API times to UTC.
type Timestamp struct {
	time.Time
}

func NewTimestamp(value time.Time) Timestamp { return Timestamp{Time: value.UTC()} }

func (timestamp Timestamp) MarshalJSON() ([]byte, error) {
	if timestamp.IsZero() {
		return nil, errors.New("timestamp must not be zero")
	}
	return json.Marshal(timestamp.UTC().Format(time.RFC3339Nano))
}

func (timestamp *Timestamp) UnmarshalJSON(data []byte) error {
	if timestamp == nil {
		return errors.New("cannot unmarshal timestamp into nil receiver")
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode timestamp: %w", err)
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return fmt.Errorf("parse RFC 3339 timestamp: %w", err)
	}
	timestamp.Time = parsed.UTC()
	return nil
}

func (timestamp Timestamp) Value() (driver.Value, error) {
	if timestamp.IsZero() {
		return nil, errors.New("timestamp must not be zero")
	}
	return timestamp.UTC(), nil
}

func (timestamp *Timestamp) Scan(src any) error {
	if timestamp == nil {
		return errors.New("cannot scan timestamp into nil receiver")
	}
	value, ok := src.(time.Time)
	if !ok {
		return fmt.Errorf("cannot scan timestamp from %T", src)
	}
	timestamp.Time = value.UTC()
	return nil
}
