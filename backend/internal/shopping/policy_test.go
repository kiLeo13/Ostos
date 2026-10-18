package shopping

import "testing"

func TestCanAccessList(t *testing.T) {
	if !CanAccessList(true) {
		t.Fatal("expected true for participant")
	}
	if CanAccessList(false) {
		t.Fatal("expected false for non-participant")
	}
}
