package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/accounts"
	"ostos/backend/internal/apierrors"
	"ostos/backend/internal/platform"
	httpmiddleware "ostos/backend/internal/transport/http/middleware"
	"ostos/backend/internal/transport/http/request"
	"ostos/backend/internal/transport/http/validation"
)

type AccountService interface {
	Register(context.Context, accounts.RegisterInput) (accounts.PublicUser, error)
	GetProfile(context.Context, platform.ID) (accounts.PublicUser, error)
	UpdateProfile(context.Context, platform.ID, accounts.UpdateProfileInput) (accounts.PublicUser, error)
	ChangePassword(context.Context, platform.ID, string, string) error
	DeleteAccount(context.Context, platform.ID, string) error
	Login(context.Context, string, string, *string) (accounts.TokenPair, error)
	Refresh(context.Context, string) (accounts.TokenPair, error)
	Logout(context.Context, accounts.AccessIdentity) error
	LogoutAll(context.Context, platform.ID) error
	VerifyEmail(context.Context, string) error
	ResendVerification(context.Context, string) error
	ForgotPassword(context.Context, string) error
	ResetPassword(context.Context, string, string) error
}

type Accounts struct {
	service   AccountService
	validator *validation.Validator
}

func NewAccounts(service AccountService, validator *validation.Validator) *Accounts {
	return &Accounts{service: service, validator: validator}
}

func (handler *Accounts) Create(context *echo.Context) error {
	var body request.UserCreateRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	user, err := handler.service.Register(context.Request().Context(), accounts.RegisterInput{
		Email: body.Email, Username: body.Username, DisplayName: body.DisplayName, Password: body.Password,
	})
	if err != nil {
		return domainError(err)
	}
	context.Response().Header().Set("Location", "/v1/users/me")
	return context.JSON(http.StatusCreated, user)
}

func (handler *Accounts) Me(context *echo.Context) error {
	identity, _ := httpmiddleware.Identity(context)
	user, err := handler.service.GetProfile(context.Request().Context(), identity.UserID)
	if err != nil {
		return domainError(err)
	}
	return context.JSON(http.StatusOK, user)
}

func (handler *Accounts) Update(context *echo.Context) error {
	var body request.UserUpdateRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	if body.Username == nil && body.DisplayName == nil {
		return apierrors.InvalidFormBody(map[string]string{"body": "Provide at least one field to update"})
	}
	identity, _ := httpmiddleware.Identity(context)
	user, err := handler.service.UpdateProfile(context.Request().Context(), identity.UserID, accounts.UpdateProfileInput{Username: body.Username, DisplayName: body.DisplayName})
	if err != nil {
		return domainError(err)
	}
	return context.JSON(http.StatusOK, user)
}

func (handler *Accounts) ChangePassword(context *echo.Context) error {
	var body request.PasswordChangeRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	if err := handler.service.ChangePassword(context.Request().Context(), identity.UserID, body.CurrentPassword, body.NewPassword); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Accounts) Delete(context *echo.Context) error {
	var body request.AccountDeleteRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	if err := handler.service.DeleteAccount(context.Request().Context(), identity.UserID, body.CurrentPassword); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Accounts) Login(context *echo.Context) error {
	var body request.LoginRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	pair, err := handler.service.Login(context.Request().Context(), body.Identifier, body.Password, body.DeviceName)
	if err != nil {
		return domainError(err)
	}
	context.Response().Header().Set("Cache-Control", "private, no-store")
	return context.JSON(http.StatusOK, pair)
}

func (handler *Accounts) Refresh(context *echo.Context) error {
	var body request.RefreshRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	pair, err := handler.service.Refresh(context.Request().Context(), body.RefreshToken)
	if err != nil {
		return domainError(err)
	}
	context.Response().Header().Set("Cache-Control", "private, no-store")
	return context.JSON(http.StatusOK, pair)
}

func (handler *Accounts) Logout(context *echo.Context) error {
	identity, _ := httpmiddleware.Identity(context)
	if err := handler.service.Logout(context.Request().Context(), identity); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Accounts) LogoutAll(context *echo.Context) error {
	identity, _ := httpmiddleware.Identity(context)
	if err := handler.service.LogoutAll(context.Request().Context(), identity.UserID); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Accounts) VerifyEmail(context *echo.Context) error {
	var body request.AccountTokenRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	if err := handler.service.VerifyEmail(context.Request().Context(), body.Token); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Accounts) ResendVerification(context *echo.Context) error {
	var body request.EmailRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	if err := handler.service.ResendVerification(context.Request().Context(), body.Email); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Accounts) ForgotPassword(context *echo.Context) error {
	var body request.EmailRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	if err := handler.service.ForgotPassword(context.Request().Context(), body.Email); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Accounts) ResetPassword(context *echo.Context) error {
	var body request.PasswordResetRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	if err := handler.service.ResetPassword(context.Request().Context(), body.Token, body.NewPassword); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}
