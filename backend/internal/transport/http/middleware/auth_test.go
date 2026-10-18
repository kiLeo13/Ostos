package middleware

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/accounts"
	"ostos/backend/internal/transport/http/response"
)

type tokenParser struct {
	identity accounts.AccessIdentity
	err      error
}

func (parser tokenParser) ParseAccessToken(string) (accounts.AccessIdentity, error) {
	return parser.identity, parser.err
}

func TestAuthenticateRequiresBearerAndStoresIdentity(t *testing.T) {
	server := echo.New()
	server.GET("/private", func(context *echo.Context) error {
		identity, ok := Identity(context)
		if !ok || identity.UserID != 11 || identity.SessionID != 22 {
			t.Fatalf("identity = %+v, %t", identity, ok)
		}
		return context.NoContent(http.StatusNoContent)
	}, Authenticate(tokenParser{identity: accounts.AccessIdentity{UserID: 11, SessionID: 22}}))

	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "bEaReR token")
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent || recorder.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("response = %d headers=%v", recorder.Code, recorder.Header())
	}
}

func TestAuthenticateRejectsMissingAndInvalidTokens(t *testing.T) {
	for _, authorization := range []string{"", "Basic token", "Bearer invalid"} {
		server := echo.New()
		server.HTTPErrorHandler = response.ErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))
		server.GET("/private", func(context *echo.Context) error { return context.NoContent(http.StatusNoContent) },
			Authenticate(tokenParser{err: errors.New("invalid")}))
		request := httptest.NewRequest(http.MethodGet, "/private", nil)
		request.Header.Set("Authorization", authorization)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("authorization %q returned %d", authorization, recorder.Code)
		}
	}
}

func TestRateLimiterBoundsOneClient(t *testing.T) {
	server := echo.New()
	server.HTTPErrorHandler = response.ErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))
	limiter := NewRateLimiter(2, time.Minute)
	server.GET("/", func(context *echo.Context) error { return context.NoContent(http.StatusNoContent) }, limiter.Middleware)
	for attempt, expected := range []int{http.StatusNoContent, http.StatusNoContent, http.StatusTooManyRequests} {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = "192.0.2.1:1234"
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		if recorder.Code != expected {
			t.Fatalf("attempt %d = %d, want %d", attempt, recorder.Code, expected)
		}
	}
}
