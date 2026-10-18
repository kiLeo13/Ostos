package middleware

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// BodyLimit bounds request bodies before decoding while preserving a typed MaxBytesError.
func BodyLimit(maxBytes int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context *echo.Context) error {
			request := context.Request()
			if request.Body != nil {
				request.Body = http.MaxBytesReader(context.Response(), request.Body, maxBytes)
			}
			return next(context)
		}
	}
}
