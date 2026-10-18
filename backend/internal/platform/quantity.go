package platform

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const maxQuantityMilli int64 = 999_999_999_999

// Quantity is an exact DECIMAL(12,3) value stored as integer thousandths.
type Quantity struct {
	milli int64
}

func ParseQuantity(value string) (Quantity, error) {
	if value == "" || strings.TrimSpace(value) != value {
		return Quantity{}, errors.New("quantity must be a plain decimal")
	}
	if strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") {
		return Quantity{}, errors.New("quantity must be nonnegative")
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return Quantity{}, errors.New("quantity must be a plain decimal")
	}
	if !isASCIIDigits(parts[0]) {
		return Quantity{}, errors.New("quantity must contain only decimal digits")
	}

	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if fraction == "" || len(fraction) > 3 || !isASCIIDigits(fraction) {
			return Quantity{}, errors.New("quantity supports at most three decimal places")
		}
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole > maxQuantityMilli/1000 {
		return Quantity{}, errors.New("quantity exceeds DECIMAL(12,3)")
	}
	for len(fraction) < 3 {
		fraction += "0"
	}
	var fractional int64
	if fraction != "" {
		fractional, _ = strconv.ParseInt(fraction, 10, 64)
	}
	milli := whole*1000 + fractional
	if milli > maxQuantityMilli {
		return Quantity{}, errors.New("quantity exceeds DECIMAL(12,3)")
	}

	return Quantity{milli: milli}, nil
}

func NewQuantityFromMilli(milli int64) (Quantity, error) {
	if milli < 0 || milli > maxQuantityMilli {
		return Quantity{}, errors.New("quantity thousandths are outside DECIMAL(12,3)")
	}
	return Quantity{milli: milli}, nil
}

func (quantity Quantity) Milli() int64 { return quantity.milli }

func (quantity Quantity) String() string {
	whole := quantity.milli / 1000
	fraction := quantity.milli % 1000
	return fmt.Sprintf("%d.%03d", whole, fraction)
}

func (quantity Quantity) MarshalJSON() ([]byte, error) { return json.Marshal(quantity.String()) }

func (quantity *Quantity) UnmarshalJSON(data []byte) error {
	if quantity == nil {
		return errors.New("cannot unmarshal quantity into nil receiver")
	}

	raw := string(data)
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("decode quantity: %w", err)
		}
	}
	parsed, err := ParseQuantity(raw)
	if err != nil {
		return err
	}
	*quantity = parsed
	return nil
}

func (quantity Quantity) Value() (driver.Value, error) { return quantity.String(), nil }

func (quantity *Quantity) Scan(src any) error {
	if quantity == nil {
		return errors.New("cannot scan quantity into nil receiver")
	}
	var raw string
	switch value := src.(type) {
	case []byte:
		raw = string(value)
	case string:
		raw = value
	default:
		return fmt.Errorf("cannot scan quantity from %T", src)
	}
	parsed, err := ParseQuantity(raw)
	if err != nil {
		return err
	}
	*quantity = parsed
	return nil
}

func isASCIIDigits(value string) bool {
	if value == "" {
		return false
	}
	for index := range len(value) {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	return true
}
