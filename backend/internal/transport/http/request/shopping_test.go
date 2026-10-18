package request

import (
	"testing"

	"ostos/backend/internal/transport/http/validation"
)

func TestShoppingRequestValidation(t *testing.T) {
	validator, err := validation.New()
	if err != nil {
		t.Fatalf("new validator: %v", err)
	}

	validCreate := ListItemCreateRequest{
		Name:     "Milk",
		Quantity: "2.000",
		Unit:     "l",
	}
	if fields := validator.Struct(validCreate); len(fields) > 0 {
		t.Fatalf("valid create failed: %#v", fields)
	}

	invalidCreate := ListItemCreateRequest{
		Name:     "   ",
		Quantity: "invalid",
		Unit:     "pounds",
	}
	fields := validator.Struct(invalidCreate)
	if fields["name"] == "" || fields["quantity"] == "" || fields["unit"] == "" {
		t.Fatalf("expected failures on name, quantity, unit: %#v", fields)
	}

	validComplete := ShoppingListCompleteRequest{
		DefaultUncheckedAction: "keep",
		UncheckedItemOverrides: []ItemActionOverrideRequest{
			{ItemID: "241895317264384000", Action: "discard"},
		},
	}
	if fields := validator.Struct(validComplete); len(fields) > 0 {
		t.Fatalf("valid complete failed: %#v", fields)
	}

	invalidComplete := ShoppingListCompleteRequest{
		DefaultUncheckedAction: "delete",
	}
	if fields := validator.Struct(invalidComplete); fields["default_unchecked_action"] == "" {
		t.Fatalf("expected failure on invalid action: %#v", fields)
	}
}
