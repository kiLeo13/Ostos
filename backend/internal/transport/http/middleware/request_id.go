package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/platform"
)

const requestIDContextKey = "ostos.request_id"

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context *echo.Context) error {
			bytes := make([]byte, 16)
			if _, err := rand.Read(bytes); err != nil {
				return err
			}
			requestID := hex.EncodeToString(bytes)
			context.Set(requestIDContextKey, requestID)
			context.SetRequest(context.Request().WithContext(platform.WithRequestID(context.Request().Context(), requestID)))
			context.Response().Header().Set("X-Request-ID", requestID)
			return next(context)
		}
	}
}

func CurrentRequestID(context *echo.Context) string {
	requestID, _ := context.Get(requestIDContextKey).(string)
	return requestID
}
