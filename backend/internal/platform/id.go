package platform

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	// SnowflakeEpoch is frozen by ADR 0001 and must never change after production use.
	SnowflakeEpoch int64 = 1767225600000 // 2026-01-01T00:00:00Z in milliseconds.
	nodeBits             = 10
	sequenceBits         = 12
	maxNodeID            = 1<<nodeBits - 1
	maxSequence          = 1<<sequenceBits - 1
)

var (
	ErrClockMovedBackwards = errors.New("clock moved backwards")
	ErrIDSpaceExhausted    = errors.New("snowflake timestamp space exhausted")
)

// ID is a positive Snowflake identifier. Its JSON representation is a decimal string.
type ID int64

func (id ID) String() string { return strconv.FormatInt(int64(id), 10) }

func (id ID) MarshalJSON() ([]byte, error) {
	if id <= 0 {
		return nil, errors.New("ID must be positive")
	}

	return json.Marshal(id.String())
}

func (id *ID) UnmarshalJSON(data []byte) error {
	if id == nil {
		return errors.New("cannot unmarshal ID into nil receiver")
	}

	if len(data) < 3 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("ID must be a decimal JSON string")
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode ID: %w", err)
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return errors.New("ID must be a positive decimal string")
	}

	*id = ID(value)
	return nil
}

// ParseID parses a positive decimal string into a Snowflake ID.
func ParseID(raw string) (ID, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, errors.New("ID must be a positive decimal string")
	}
	return ID(value), nil
}

func (id ID) Value() (driver.Value, error) {
	if id <= 0 {
		return nil, errors.New("ID must be positive")
	}
	return int64(id), nil
}

func (id *ID) Scan(src any) error {
	if id == nil {
		return errors.New("cannot scan ID into nil receiver")
	}

	var raw string
	switch value := src.(type) {
	case int64:
		raw = strconv.FormatInt(value, 10)
	case []byte:
		raw = string(bytes.Clone(value))
	case string:
		raw = value
	default:
		return fmt.Errorf("cannot scan ID from %T", src)
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return errors.New("database ID must be positive")
	}
	*id = ID(value)
	return nil
}

// Clock supplies time to generators and services without coupling tests to wall time.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// SystemClock supplies the current wall clock to application services.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// IDGenerator creates entity identifiers.
type IDGenerator interface {
	Next(context.Context) (ID, error)
}

// SnowflakeGenerator is safe for concurrent use by one API process.
type SnowflakeGenerator struct {
	mu         sync.Mutex
	nodeID     int64
	lastMillis int64
	sequence   int64
	clock      Clock
	wait       func(context.Context) error
}

func NewSnowflakeGenerator(nodeID int64) (*SnowflakeGenerator, error) {
	return newSnowflakeGenerator(nodeID, systemClock{}, waitOneMillisecond)
}

func newSnowflakeGenerator(nodeID int64, clock Clock, wait func(context.Context) error) (*SnowflakeGenerator, error) {
	if nodeID < 0 || nodeID > maxNodeID {
		return nil, fmt.Errorf("snowflake node ID must be between 0 and %d", maxNodeID)
	}
	if clock == nil || wait == nil {
		return nil, errors.New("snowflake clock and wait function are required")
	}

	return &SnowflakeGenerator{nodeID: nodeID, clock: clock, wait: wait}, nil
}

func (generator *SnowflakeGenerator) Next(ctx context.Context) (ID, error) {
	generator.mu.Lock()
	defer generator.mu.Unlock()

	for {
		millis := generator.clock.Now().UTC().UnixMilli() - SnowflakeEpoch
		if millis < 0 {
			return 0, ErrClockMovedBackwards
		}
		if millis >= 1<<41 {
			return 0, ErrIDSpaceExhausted
		}
		if millis < generator.lastMillis {
			return 0, ErrClockMovedBackwards
		}

		if millis == generator.lastMillis {
			generator.sequence = (generator.sequence + 1) & maxSequence
			if generator.sequence == 0 {
				if err := generator.wait(ctx); err != nil {
					return 0, err
				}
				continue
			}
		} else {
			generator.sequence = 0
			generator.lastMillis = millis
		}

		value := (millis << (nodeBits + sequenceBits)) | (generator.nodeID << sequenceBits) | generator.sequence
		if value <= 0 {
			return 0, ErrIDSpaceExhausted
		}
		return ID(value), nil
	}
}

func waitOneMillisecond(ctx context.Context) error {
	timer := time.NewTimer(time.Millisecond)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
