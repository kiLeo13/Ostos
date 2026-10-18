package router

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/transport/http/handler"
	httpmiddleware "ostos/backend/internal/transport/http/middleware"
	"ostos/backend/internal/transport/http/response"
)

type Dependencies struct {
	Logger    *slog.Logger
	Database  handler.Pinger
	BodyLimit int64
}

func New(dependencies Dependencies) *echo.Echo {
	server := echo.New()
	server.Logger = dependencies.Logger
	server.HTTPErrorHandler = response.ErrorHandler(dependencies.Logger)
	server.Use(httpmiddleware.BodyLimit(dependencies.BodyLimit))

	health := handler.NewHealth(dependencies.Database, 2*time.Second)
	server.GET("/health/live", health.Live)
	server.GET("/health/ready", health.Ready)
	return server
}
