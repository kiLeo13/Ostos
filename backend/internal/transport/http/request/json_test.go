package request_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/apierrors"
	httpmiddleware "ostos/backend/internal/transport/http/middleware"
	"ostos/backend/internal/transport/http/request"
	"ostos/backend/internal/transport/http/response"
	"ostos/backend/internal/transport/http/validation"
)

type testRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=2,max=32,username"`
}

func TestDecodeJSONAndErrorEnvelope(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		bodyLimit   int64
		status      int
		code        string
		fieldCount  int
	}{
		{name: "valid", contentType: "application/json; charset=utf-8", body: `{"email":"user@example.com","username":"user_name"}`, bodyLimit: 1024, status: http.StatusNoContent},
		{name: "unknown field", contentType: "application/json", body: `{"email":"user@example.com","username":"user_name","admin":true}`, bodyLimit: 1024, status: http.StatusBadRequest, code: apierrors.CodeMalformedJSON},
		{name: "multiple values", contentType: "application/json", body: `{}` + `{}`, bodyLimit: 1024, status: http.StatusBadRequest, code: apierrors.CodeMalformedJSON},
		{name: "malformed", contentType: "application/json", body: `{`, bodyLimit: 1024, status: http.StatusBadRequest, code: apierrors.CodeMalformedJSON},
		{name: "wrong media type", contentType: "text/plain", body: `{}`, bodyLimit: 1024, status: http.StatusUnsupportedMediaType, code: apierrors.CodeUnsupportedMediaType},
		{name: "validation collects fields", contentType: "application/json", body: `{"email":"bad","username":"Upper Case"}`, bodyLimit: 1024, status: http.StatusUnprocessableEntity, code: apierrors.CodeInvalidFormBody, fieldCount: 2},
		{name: "body limit", contentType: "application/json", body: `{"email":"user@example.com","username":"user_name"}`, bodyLimit: 8, status: http.StatusRequestEntityTooLarge, code: apierrors.CodeRequestTooLarge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator, err := validation.New()
			if err != nil {
				t.Fatalf("validator: %v", err)
			}
			server := echo.New()
			server.HTTPErrorHandler = response.ErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))
			server.Use(httpmiddleware.BodyLimit(test.bodyLimit))
			server.POST("/test", func(context *echo.Context) error {
				var destination testRequest
				if err := request.DecodeJSON(context.Request(), &destination, validator); err != nil {
					return err
				}
				return context.NoContent(http.StatusNoContent)
			})

			httpRequest := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			httpRequest.Header.Set("Content-Type", test.contentType)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, httpRequest)
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, test.status, recorder.Body.String())
			}
			if test.code == "" {
				return
			}
			var envelope apierrors.Envelope
			if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode envelope: %v", err)
			}
			if envelope.Errors.Code != test.code {
				t.Fatalf("code = %q, want %q", envelope.Errors.Code, test.code)
			}
			if len(envelope.Errors.Fields) != test.fieldCount {
				t.Fatalf("fields = %#v", envelope.Errors.Fields)
			}
		})
	}
}
