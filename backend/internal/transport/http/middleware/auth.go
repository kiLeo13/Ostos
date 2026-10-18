package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/accounts"
	"ostos/backend/internal/apierrors"
)

const identityContextKey = "ostos.auth.identity"

type AccessTokenParser interface {
	ParseAccessToken(string) (accounts.AccessIdentity, error)
}

func Authenticate(parser AccessTokenParser) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context *echo.Context) error {
			parts := strings.Fields(context.Request().Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return apierrors.New(http.StatusUnauthorized, "INVALID_AUTHENTICATION", "Valid authentication is required")
			}
			identity, err := parser.ParseAccessToken(parts[1])
			if err != nil {
				return apierrors.New(http.StatusUnauthorized, "INVALID_AUTHENTICATION", "Valid authentication is required")
			}
			context.Set(identityContextKey, identity)
			context.Response().Header().Set("Cache-Control", "private, no-store")
			return next(context)
		}
	}
}

func Identity(context *echo.Context) (accounts.AccessIdentity, bool) {
	identity, ok := context.Get(identityContextKey).(accounts.AccessIdentity)
	return identity, ok
}
