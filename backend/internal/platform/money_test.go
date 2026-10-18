package platform

import (
	"encoding/json"
	"testing"
)

func TestCentsRequireNonnegativeInteger(t *testing.T) {
	for _, input := range []string{"0", "1299"} {
		var cents Cents
		if err := json.Unmarshal([]byte(input), &cents); err != nil {
			t.Fatalf("unmarshal %s: %v", input, err)
		}
	}
	for _, input := range []string{"-1", "1.2", `"12"`} {
		var cents Cents
		if err := json.Unmarshal([]byte(input), &cents); err == nil {
			t.Fatalf("invalid cents %s were accepted", input)
		}
	}
	if _, err := NewCents(-1); err == nil {
		t.Fatal("negative cents were accepted")
	}
}
