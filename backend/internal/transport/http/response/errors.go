package response

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/apierrors"
)

func ErrorHandler(logger *slog.Logger) echo.HTTPErrorHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return func(context *echo.Context, err error) {
		public := mapEchoError(err)
		if public.Status >= http.StatusInternalServerError {
			logger.ErrorContext(context.Request().Context(), "request failed", "status", public.Status, "code", public.Code, "error", err)
		}
		if writeErr := context.JSON(public.Status, apierrors.ToEnvelope(public)); writeErr != nil {
			logger.ErrorContext(context.Request().Context(), "write error response", "error", writeErr)
		}
	}
}

func mapEchoError(err error) *apierrors.Error {
	var echoError *echo.HTTPError
	if errors.As(err, &echoError) {
		switch echoError.Code {
		case http.StatusNotFound:
			return apierrors.New(http.StatusNotFound, apierrors.CodeNotFound, "The requested resource was not found")
		case http.StatusMethodNotAllowed:
			return apierrors.New(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "The request method is not allowed")
		}
	}
	return apierrors.Public(err)
}
