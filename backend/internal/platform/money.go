package platform

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Cents is a nonnegative BRL amount in the smallest currency unit.
type Cents int64

func NewCents(value int64) (Cents, error) {
	if value < 0 {
		return 0, errors.New("cents must be nonnegative")
	}
	return Cents(value), nil
}

func (cents Cents) MarshalJSON() ([]byte, error) {
	if cents < 0 {
		return nil, errors.New("cents must be nonnegative")
	}
	return json.Marshal(int64(cents))
}

func (cents *Cents) UnmarshalJSON(data []byte) error {
	if cents == nil {
		return errors.New("cannot unmarshal cents into nil receiver")
	}
	value, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil || value < 0 {
		return errors.New("cents must be a nonnegative integer")
	}
	*cents = Cents(value)
	return nil
}

func (cents Cents) Value() (driver.Value, error) {
	if cents < 0 {
		return nil, errors.New("cents must be nonnegative")
	}
	return int64(cents), nil
}

func (cents *Cents) Scan(src any) error {
	if cents == nil {
		return errors.New("cannot scan cents into nil receiver")
	}
	value, ok := src.(int64)
	if !ok {
		return fmt.Errorf("cannot scan cents from %T", src)
	}
	parsed, err := NewCents(value)
	if err != nil {
		return err
	}
	*cents = parsed
	return nil
}
