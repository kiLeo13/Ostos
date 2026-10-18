package router

import (
	"log/slog"
	"testing"

	"ostos/backend/internal/transport/http/handler"
	httpmiddleware "ostos/backend/internal/transport/http/middleware"
	"ostos/backend/internal/transport/http/validation"
)

type routeServiceStub struct {
	handler.AccountService
	handler.FamilyService
	handler.ShoppingService
	httpmiddleware.AccessTokenParser
}

func TestRouterPublishesEveryStepTwoThreeAndFourRoute(t *testing.T) {
	validator, err := validation.New()
	if err != nil {
		t.Fatalf("validator: %v", err)
	}
	stub := &routeServiceStub{}
	server := New(Dependencies{
		Logger: slog.Default(), BodyLimit: 1024, Accounts: stub, AccessTokens: stub, Families: stub, Shopping: stub, Validator: validator,
	})
	expected := map[string]bool{
		"POST /v1/users": false, "GET /v1/users/me": false, "PATCH /v1/users/me": false,
		"PUT /v1/users/me/password": false, "DELETE /v1/users/me": false,
		"POST /v1/auth/login": false, "POST /v1/auth/refresh": false, "POST /v1/auth/logout": false,
		"POST /v1/auth/logout-all": false, "POST /v1/auth/verify-email": false,
		"POST /v1/auth/resend-verification": false, "POST /v1/auth/forgot-password": false,
		"POST /v1/auth/reset-password": false,
		"GET /v1/families":             false, "POST /v1/families": false, "GET /v1/families/:family_id": false,
		"PATCH /v1/families/:family_id": false, "DELETE /v1/families/:family_id": false,
		"GET /v1/families/:family_id/members":               false,
		"PATCH /v1/families/:family_id/members/:member_id":  false,
		"DELETE /v1/families/:family_id/members/:member_id": false,
		"POST /v1/families/:family_id/leave":                false,
		"POST /v1/families/:family_id/transfer-ownership":   false,
		"GET /v1/invitations":                               false, "POST /v1/invitations/:invitation_id/accept": false,
		"POST /v1/invitations/:invitation_id/decline":               false,
		"GET /v1/families/:family_id/invitations":                   false,
		"POST /v1/families/:family_id/invitations":                  false,
		"DELETE /v1/families/:family_id/invitations/:invitation_id": false,
		"GET /v1/families/:family_id/shopping-list":                  false,
		"POST /v1/families/:family_id/shopping-list/items":           false,
		"PATCH /v1/families/:family_id/shopping-list/items/:item_id":  false,
		"DELETE /v1/families/:family_id/shopping-list/items/:item_id": false,
		"PUT /v1/families/:family_id/shopping-list/items/order":       false,
		"POST /v1/families/:family_id/shopping-list/complete":         false,
		"GET /v1/families/:family_id/shopping-trips":                 false,
		"GET /v1/families/:family_id/shopping-trips/:trip_id":        false,
	}
	for _, route := range server.Router().Routes() {
		key := route.Method + " " + route.Path
		if _, exists := expected[key]; exists {
			expected[key] = true
		}
	}
	for route, found := range expected {
		if !found {
			t.Errorf("route %s is not registered", route)
		}
	}
}
