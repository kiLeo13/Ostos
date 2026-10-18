package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/apierrors"
)

type rateWindow struct {
	count    int
	resetsAt time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]rateWindow
	limit   int
	window  time.Duration
	clock   func() time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{clients: make(map[string]rateWindow), limit: limit, window: window, clock: time.Now}
}

func (limiter *RateLimiter) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(context *echo.Context) error {
		key := clientAddress(context.Request().RemoteAddr)
		now := limiter.clock()
		limiter.mu.Lock()
		entry := limiter.clients[key]
		if entry.resetsAt.IsZero() || !now.Before(entry.resetsAt) {
			entry = rateWindow{resetsAt: now.Add(limiter.window)}
		}
		entry.count++
		limiter.clients[key] = entry
		limited := entry.count > limiter.limit
		limiter.mu.Unlock()
		if limited {
			return apierrors.New(http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests; try again later")
		}
		return next(context)
	}
}

func clientAddress(remote string) string {
	host, _, err := net.SplitHostPort(remote)
	if err == nil {
		return host
	}
	return remote
}
