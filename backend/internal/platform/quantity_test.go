package platform

import (
	"encoding/json"
	"testing"
)

func TestQuantityParsesExactThousandths(t *testing.T) {
	tests := map[string]int64{
		"0":             0,
		"1":             1000,
		"1.2":           1200,
		"1.23":          1230,
		"1.234":         1234,
		"999999999.999": maxQuantityMilli,
	}
	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			quantity, err := ParseQuantity(input)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if quantity.Milli() != expected {
				t.Fatalf("milli = %d, want %d", quantity.Milli(), expected)
			}
		})
	}
}

func TestQuantityRejectsNonDecimalAndOverflow(t *testing.T) {
	for _, input := range []string{"", " 1", "+1", "-1", ".1", "1.", "1.0000", "1e3", "NaN", "1000000000"} {
		if _, err := ParseQuantity(input); err == nil {
			t.Fatalf("invalid quantity %q was accepted", input)
		}
	}
}

func TestQuantityJSONNeverUsesFloatingPoint(t *testing.T) {
	for _, input := range []string{`"1.230"`, `1.230`} {
		var quantity Quantity
		if err := json.Unmarshal([]byte(input), &quantity); err != nil {
			t.Fatalf("unmarshal %s: %v", input, err)
		}
		if quantity.Milli() != 1230 {
			t.Fatalf("milli = %d", quantity.Milli())
		}
		encoded, err := json.Marshal(quantity)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(encoded) != `"1.230"` {
			t.Fatalf("encoded = %s", encoded)
		}
	}
}
