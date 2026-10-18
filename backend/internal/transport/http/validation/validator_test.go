package validation

import (
	"testing"

	validator "github.com/go-playground/validator/v10"
)

type registrationRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=2,max=32,username"`
}

func TestValidatorCollectsEveryFieldFailure(t *testing.T) {
	instance, err := New()
	if err != nil {
		t.Fatalf("new validator: %v", err)
	}
	fields := instance.Struct(registrationRequest{Email: "invalid", Username: "Upper Case"})
	if len(fields) != 2 {
		t.Fatalf("fields = %#v", fields)
	}
	if fields["email"] != "Enter a valid email address" {
		t.Fatalf("email message = %q", fields["email"])
	}
	if fields["username"] != "Use 2-32 lowercase letters, digits, hyphens, or underscores" {
		t.Fatalf("username message = %q", fields["username"])
	}
}

func TestUsernameValidatorBoundariesAndCharacters(t *testing.T) {
	instance, _ := New()
	valid := []string{"ab", "a-b_c9", "abcdefghijklmnopqrstuvwxyz123456"}
	for _, username := range valid {
		if fields := instance.Struct(registrationRequest{Email: "user@example.com", Username: username}); len(fields) > 0 {
			t.Fatalf("valid username %q: %#v", username, fields)
		}
	}
	invalid := []string{"a", "abcdefghijklmnopqrstuvwxyz1234567", "Upper", "with space", "usuário", "bad.dot", "bad!"}
	for _, username := range invalid {
		if fields := instance.Struct(registrationRequest{Email: "user@example.com", Username: username}); fields["username"] == "" {
			t.Fatalf("invalid username %q was accepted", username)
		}
	}
}

func TestEngineSupportsStructLevelValidation(t *testing.T) {
	type choice struct {
		Username string `json:"username"`
		UserID   string `json:"user_id"`
	}
	instance, _ := New()
	instance.Engine().RegisterStructValidation(func(level validator.StructLevel) {
		value := level.Current().Interface().(choice)
		if (value.Username == "") == (value.UserID == "") {
			level.ReportError(value.Username, "username", "username", "exactlyone", "")
		}
	}, choice{})
	if fields := instance.Struct(choice{}); fields["username"] == "" {
		t.Fatalf("struct-level failure missing: %#v", fields)
	}
	if fields := instance.Struct(choice{Username: "exact"}); len(fields) != 0 {
		t.Fatalf("valid related fields failed: %#v", fields)
	}
}
