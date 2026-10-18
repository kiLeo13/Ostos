package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/apierrors"
)

func TestRequestLoggerRecordsSafeStructuredFields(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	server := echo.New()
	server.Use(RequestID(), RequestLogger(logger))
	server.GET("/failure", func(*echo.Context) error {
		return apierrors.New(http.StatusConflict, "STATE_CONFLICT", "Conflict")
	})

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/failure", nil))
	logged := output.String()
	for _, expected := range []string{`"route":"/failure"`, `"status":409`, `"error_code":"STATE_CONFLICT"`, `"request_id":`} {
		if !strings.Contains(logged, expected) {
			t.Fatalf("log %q does not contain %q", logged, expected)
		}
	}
	if strings.Contains(logged, "Conflict") {
		t.Fatalf("log contains application error message: %q", logged)
	}
}
