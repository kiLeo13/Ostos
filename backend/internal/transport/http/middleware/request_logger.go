package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/apierrors"
)

func RequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context *echo.Context) error {
			started := time.Now()
			err := next(context)
			_, status := echo.ResolveResponseStatus(context.Response(), err)
			code := ""
			if err != nil {
				public := apierrors.Public(err)
				status, code = public.Status, public.Code
			}
			if status == 0 {
				status = http.StatusOK
			}
			attributes := []any{
				"request_id", CurrentRequestID(context), "method", context.Request().Method,
				"route", context.Path(), "status", status, "duration_ms", time.Since(started).Milliseconds(),
				"client_ip", context.RealIP(),
			}
			if code != "" {
				attributes = append(attributes, "error_code", code)
			}
			if identity, ok := Identity(context); ok {
				attributes = append(attributes, "actor_id", identity.UserID.String())
			}
			logger.InfoContext(context.Request().Context(), "HTTP request", attributes...)
			return err
		}
	}
}
