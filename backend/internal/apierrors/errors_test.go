package apierrors

import (
	"errors"
	"net/http"
	"testing"
)

func TestPublicHidesUnknownErrors(t *testing.T) {
	public := Public(errors.New("database password leaked here"))
	if public.Status != http.StatusInternalServerError || public.Code != CodeInternal {
		t.Fatalf("unexpected public error: %+v", public)
	}
	if public.Message != "An internal error occurred" {
		t.Fatalf("unsafe message: %q", public.Message)
	}
}

func TestEnvelopeOmitsFieldsWhenUnavailable(t *testing.T) {
	envelope := ToEnvelope(MalformedJSON(errors.New("syntax")))
	if envelope.Errors.Fields != nil {
		t.Fatalf("malformed JSON fields = %#v", envelope.Errors.Fields)
	}
	validation := ToEnvelope(InvalidFormBody(map[string]string{"email": "invalid"}))
	if validation.Errors.Fields["email"] != "invalid" {
		t.Fatalf("validation fields = %#v", validation.Errors.Fields)
	}
}
