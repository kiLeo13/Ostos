package handler_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ostos/backend/internal/transport/http/response"
	"ostos/backend/internal/transport/http/router"
)

type pinger struct{ err error }

func (instance pinger) PingContext(context.Context) error { return instance.err }

func TestHealthRoutes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, test := range []struct {
		name   string
		path   string
		pinger pinger
		status int
	}{
		{name: "healthy", path: "/health", pinger: pinger{}, status: http.StatusOK},
		{name: "database unavailable", path: "/health", pinger: pinger{err: errors.New("down")}, status: http.StatusServiceUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := router.New(router.Dependencies{Logger: logger, Database: test.pinger, BodyLimit: 1024})
			server.HTTPErrorHandler = response.ErrorHandler(logger)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}
