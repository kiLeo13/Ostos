package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/apierrors"
)

type Pinger interface {
	PingContext(context.Context) error
}

type Health struct {
	database Pinger
	timeout  time.Duration
}

func NewHealth(database Pinger, timeout time.Duration) *Health {
	return &Health{database: database, timeout: timeout}
}

func (health *Health) Check(echoContext *echo.Context) error {
	if health.database == nil {
		return apierrors.New(http.StatusServiceUnavailable, apierrors.CodeUnavailable, "A required dependency is unavailable")
	}

	ctx, cancel := context.WithTimeout(echoContext.Request().Context(), health.timeout)
	defer cancel()
	if err := health.database.PingContext(ctx); err != nil {
		return &apierrors.Error{
			Status:  http.StatusServiceUnavailable,
			Code:    apierrors.CodeUnavailable,
			Message: "A required dependency is unavailable",
			Cause:   err,
		}
	}
	return echoContext.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
