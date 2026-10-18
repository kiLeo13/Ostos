package router

import (
	"log/slog"
	"net/netip"
	"time"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
	"ostos/backend/internal/transport/http/handler"
	httpmiddleware "ostos/backend/internal/transport/http/middleware"
	"ostos/backend/internal/transport/http/response"
	"ostos/backend/internal/transport/http/validation"
)

type Dependencies struct {
	Logger         *slog.Logger
	Database       handler.Pinger
	BodyLimit      int64
	Accounts       handler.AccountService
	AccessTokens   httpmiddleware.AccessTokenParser
	Families       handler.FamilyService
	Shopping       handler.ShoppingService
	Validator      *validation.Validator
	CORSOrigins    []string
	TrustedProxies []netip.Prefix
}

func New(dependencies Dependencies) *echo.Echo {
	server := echo.New()
	server.Logger = dependencies.Logger
	server.HTTPErrorHandler = response.ErrorHandler(dependencies.Logger)
	server.IPExtractor = httpmiddleware.TrustedProxyIPExtractor(dependencies.TrustedProxies)
	server.Use(httpmiddleware.BodyLimit(dependencies.BodyLimit))
	server.Use(httpmiddleware.RequestID())
	server.Use(httpmiddleware.RequestLogger(dependencies.Logger))
	if len(dependencies.CORSOrigins) > 0 {
		cors, err := (echomiddleware.CORSConfig{
			AllowOrigins:  dependencies.CORSOrigins,
			AllowMethods:  []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:  []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
			ExposeHeaders: []string{"Location", "X-Request-ID"},
			MaxAge:        600,
		}).ToMiddleware()
		if err != nil {
			panic("invalid CORS configuration: " + err.Error())
		}
		server.Use(cors)
	}

	health := handler.NewHealth(dependencies.Database, 2*time.Second)
	server.GET("/health", health.Check)

	if dependencies.Accounts == nil || dependencies.AccessTokens == nil || dependencies.Families == nil || dependencies.Validator == nil {
		return server
	}
	accountHandler := handler.NewAccounts(dependencies.Accounts, dependencies.Validator)
	familyHandler := handler.NewFamilies(dependencies.Families, dependencies.Validator)
	authenticate := httpmiddleware.Authenticate(dependencies.AccessTokens)
	authRateLimit := httpmiddleware.NewRateLimiter(10, time.Minute).Middleware

	server.POST("/v1/users", accountHandler.Create, authRateLimit)
	server.POST("/v1/auth/login", accountHandler.Login, authRateLimit)
	server.POST("/v1/auth/refresh", accountHandler.Refresh, authRateLimit)
	server.POST("/v1/auth/verify-email", accountHandler.VerifyEmail, authRateLimit)
	server.POST("/v1/auth/resend-verification", accountHandler.ResendVerification, authRateLimit)
	server.POST("/v1/auth/forgot-password", accountHandler.ForgotPassword, authRateLimit)
	server.POST("/v1/auth/reset-password", accountHandler.ResetPassword, authRateLimit)

	server.GET("/v1/users/me", accountHandler.Me, authenticate)
	server.PATCH("/v1/users/me", accountHandler.Update, authenticate)
	server.PUT("/v1/users/me/password", accountHandler.ChangePassword, authenticate)
	server.DELETE("/v1/users/me", accountHandler.Delete, authenticate)
	server.POST("/v1/auth/logout", accountHandler.Logout, authenticate)
	server.POST("/v1/auth/logout-all", accountHandler.LogoutAll, authenticate)

	server.GET("/v1/families", familyHandler.List, authenticate)
	server.POST("/v1/families", familyHandler.Create, authenticate)
	server.GET("/v1/families/:family_id", familyHandler.Get, authenticate)
	server.PATCH("/v1/families/:family_id", familyHandler.Update, authenticate)
	server.DELETE("/v1/families/:family_id", familyHandler.Delete, authenticate)
	server.GET("/v1/families/:family_id/members", familyHandler.Members, authenticate)
	server.PATCH("/v1/families/:family_id/members/:member_id", familyHandler.ChangeRole, authenticate)
	server.DELETE("/v1/families/:family_id/members/:member_id", familyHandler.RemoveMember, authenticate)
	server.POST("/v1/families/:family_id/leave", familyHandler.Leave, authenticate)
	server.POST("/v1/families/:family_id/transfer-ownership", familyHandler.TransferOwnership, authenticate)
	server.GET("/v1/invitations", familyHandler.ReceivedInvitations, authenticate)
	server.POST("/v1/invitations/:invitation_id/accept", familyHandler.AcceptInvitation, authenticate)
	server.POST("/v1/invitations/:invitation_id/decline", familyHandler.DeclineInvitation, authenticate)
	server.GET("/v1/families/:family_id/invitations", familyHandler.FamilyInvitations, authenticate)
	server.POST("/v1/families/:family_id/invitations", familyHandler.Invite, authenticate)
	server.DELETE("/v1/families/:family_id/invitations/:invitation_id", familyHandler.CancelInvitation, authenticate)

	if dependencies.Shopping != nil {
		shoppingHandler := handler.NewShopping(dependencies.Shopping, dependencies.Validator)
		server.GET("/v1/families/:family_id/shopping-list", shoppingHandler.List, authenticate)
		server.POST("/v1/families/:family_id/shopping-list/items", shoppingHandler.CreateItem, authenticate)
		server.PATCH("/v1/families/:family_id/shopping-list/items/:item_id", shoppingHandler.UpdateItem, authenticate)
		server.DELETE("/v1/families/:family_id/shopping-list/items/:item_id", shoppingHandler.DeleteItem, authenticate)
		server.PUT("/v1/families/:family_id/shopping-list/items/order", shoppingHandler.ReorderItems, authenticate)
		server.POST("/v1/families/:family_id/shopping-list/complete", shoppingHandler.Complete, authenticate)
		server.GET("/v1/families/:family_id/shopping-trips", shoppingHandler.ListTrips, authenticate)
		server.GET("/v1/families/:family_id/shopping-trips/:trip_id", shoppingHandler.GetTrip, authenticate)
	}
	return server
}
