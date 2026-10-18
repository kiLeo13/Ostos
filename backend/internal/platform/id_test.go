package platform

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

type fixedClock struct {
	mu  sync.Mutex
	now time.Time
}

func (clock *fixedClock) Now() time.Time {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	return clock.now
}

func (clock *fixedClock) set(value time.Time) {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	clock.now = value
}

func TestSnowflakeGeneratorProducesOrderedUniqueIDs(t *testing.T) {
	clock := &fixedClock{now: time.UnixMilli(SnowflakeEpoch + 1234)}
	generator, err := newSnowflakeGenerator(17, clock, func(context.Context) error { return nil })
	if err != nil {
		t.Fatalf("new generator: %v", err)
	}

	first, err := generator.Next(context.Background())
	if err != nil {
		t.Fatalf("first ID: %v", err)
	}
	second, err := generator.Next(context.Background())
	if err != nil {
		t.Fatalf("second ID: %v", err)
	}
	if second <= first {
		t.Fatalf("IDs are not ordered: first=%d second=%d", first, second)
	}
	if node := (int64(first) >> sequenceBits) & maxNodeID; node != 17 {
		t.Fatalf("node bits = %d, want 17", node)
	}
}

func TestSnowflakeGeneratorRejectsInvalidNodeAndClockRegression(t *testing.T) {
	for _, node := range []int64{-1, 1024} {
		if _, err := NewSnowflakeGenerator(node); err == nil {
			t.Fatalf("node %d was accepted", node)
		}
	}

	clock := &fixedClock{now: time.UnixMilli(SnowflakeEpoch + 2)}
	generator, _ := newSnowflakeGenerator(0, clock, func(context.Context) error { return nil })
	if _, err := generator.Next(context.Background()); err != nil {
		t.Fatalf("initial ID: %v", err)
	}
	clock.set(time.UnixMilli(SnowflakeEpoch + 1))
	if _, err := generator.Next(context.Background()); !errors.Is(err, ErrClockMovedBackwards) {
		t.Fatalf("clock regression error = %v", err)
	}
}

func TestIDJSONUsesPositiveDecimalString(t *testing.T) {
	original := ID(123456789)
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal ID: %v", err)
	}
	if string(encoded) != `"123456789"` {
		t.Fatalf("encoded ID = %s", encoded)
	}

	var decoded ID
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal ID: %v", err)
	}
	if decoded != original {
		t.Fatalf("decoded ID = %d", decoded)
	}
	for _, invalid := range []string{`123`, `"0"`, `"-1"`, `"abc"`} {
		if err := json.Unmarshal([]byte(invalid), &decoded); err == nil {
			t.Fatalf("invalid ID %s was accepted", invalid)
		}
	}
}
