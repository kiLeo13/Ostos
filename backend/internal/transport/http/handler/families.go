package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/apierrors"
	"ostos/backend/internal/families"
	"ostos/backend/internal/platform"
	httpmiddleware "ostos/backend/internal/transport/http/middleware"
	"ostos/backend/internal/transport/http/request"
	"ostos/backend/internal/transport/http/validation"
)

type FamilyService interface {
	List(context.Context, platform.ID) ([]families.Summary, error)
	Create(context.Context, platform.ID, string, string) (families.Family, error)
	Get(context.Context, platform.ID, platform.ID) (families.Family, error)
	Update(context.Context, platform.ID, platform.ID, string, string) (families.Family, error)
	Delete(context.Context, platform.ID, platform.ID, string) error
	Members(context.Context, platform.ID, platform.ID) ([]families.MemberView, error)
	ChangeRole(context.Context, platform.ID, platform.ID, platform.ID, families.Role, string) error
	RemoveMember(context.Context, platform.ID, platform.ID, platform.ID, string) error
	Leave(context.Context, platform.ID, platform.ID, string) error
	Invite(context.Context, platform.ID, platform.ID, families.InviteTarget, string) error
	ReceivedInvitations(context.Context, platform.ID) ([]families.InvitationView, error)
	FamilyInvitations(context.Context, platform.ID, platform.ID) ([]families.SentInvitationView, error)
	AcceptInvitation(context.Context, platform.ID, platform.ID, string) error
	DeclineInvitation(context.Context, platform.ID, platform.ID, string) error
	CancelInvitation(context.Context, platform.ID, platform.ID, platform.ID, string) error
	TransferOwnership(context.Context, platform.ID, platform.ID, platform.ID, string) error
}

type Families struct {
	service   FamilyService
	validator *validation.Validator
}

func NewFamilies(service FamilyService, validator *validation.Validator) *Families {
	return &Families{service: service, validator: validator}
}

func (handler *Families) List(context *echo.Context) error {
	identity, _ := httpmiddleware.Identity(context)
	result, err := handler.service.List(context.Request().Context(), identity.UserID)
	if err != nil {
		return domainError(err)
	}
	if result == nil {
		result = []families.Summary{}
	}
	return context.JSON(http.StatusOK, result)
}

func (handler *Families) Create(context *echo.Context) error {
	var body request.FamilyCreateRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	family, err := handler.service.Create(context.Request().Context(), identity.UserID, body.Name, httpmiddleware.CurrentRequestID(context))
	if err != nil {
		return domainError(err)
	}
	context.Response().Header().Set("Location", "/v1/families/"+family.ID.String())
	return context.JSON(http.StatusCreated, family)
}

func (handler *Families) Get(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	family, serviceErr := handler.service.Get(context.Request().Context(), identity.UserID, familyID)
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	return context.JSON(http.StatusOK, family)
}

func (handler *Families) Update(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	var body request.FamilyUpdateRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	family, serviceErr := handler.service.Update(context.Request().Context(), identity.UserID, familyID, body.Name, httpmiddleware.CurrentRequestID(context))
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	return context.JSON(http.StatusOK, family)
}

func (handler *Families) Delete(echoContext *echo.Context) error {
	return handler.familyCommand(echoContext, func(ctx context.Context, actorID, familyID platform.ID, requestID string) error {
		return handler.service.Delete(ctx, actorID, familyID, requestID)
	})
}

func (handler *Families) Members(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	members, serviceErr := handler.service.Members(context.Request().Context(), identity.UserID, familyID)
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	if members == nil {
		members = []families.MemberView{}
	}
	return context.JSON(http.StatusOK, members)
}

func (handler *Families) ChangeRole(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	memberID, err := routeID(context, "member_id")
	if err != nil {
		return err
	}
	var body request.MemberRoleUpdateRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	if err := handler.service.ChangeRole(context.Request().Context(), identity.UserID, familyID, memberID, families.Role(body.Role), httpmiddleware.CurrentRequestID(context)); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Families) RemoveMember(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	memberID, err := routeID(context, "member_id")
	if err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	if err := handler.service.RemoveMember(context.Request().Context(), identity.UserID, familyID, memberID, httpmiddleware.CurrentRequestID(context)); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Families) Leave(echoContext *echo.Context) error {
	return handler.familyCommand(echoContext, func(ctx context.Context, actorID, familyID platform.ID, requestID string) error {
		return handler.service.Leave(ctx, actorID, familyID, requestID)
	})
}

func (handler *Families) TransferOwnership(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	var body request.TransferOwnershipRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	newOwnerID, _ := parseID(body.NewOwnerID)
	identity, _ := httpmiddleware.Identity(context)
	if err := handler.service.TransferOwnership(context.Request().Context(), identity.UserID, familyID, newOwnerID, httpmiddleware.CurrentRequestID(context)); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Families) Invite(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	var body request.InvitationCreateRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	target := families.InviteTarget{Username: body.Username}
	if body.UserID != nil {
		value, _ := parseID(*body.UserID)
		target.UserID = &value
	}
	identity, _ := httpmiddleware.Identity(context)
	if err := handler.service.Invite(context.Request().Context(), identity.UserID, familyID, target, httpmiddleware.CurrentRequestID(context)); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Families) ReceivedInvitations(context *echo.Context) error {
	identity, _ := httpmiddleware.Identity(context)
	invitations, err := handler.service.ReceivedInvitations(context.Request().Context(), identity.UserID)
	if err != nil {
		return domainError(err)
	}
	if invitations == nil {
		invitations = []families.InvitationView{}
	}
	return context.JSON(http.StatusOK, invitations)
}

func (handler *Families) FamilyInvitations(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	invitations, serviceErr := handler.service.FamilyInvitations(context.Request().Context(), identity.UserID, familyID)
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	if invitations == nil {
		invitations = []families.SentInvitationView{}
	}
	return context.JSON(http.StatusOK, invitations)
}

func (handler *Families) AcceptInvitation(context *echo.Context) error {
	return handler.invitationCommand(context, handler.service.AcceptInvitation)
}
func (handler *Families) DeclineInvitation(context *echo.Context) error {
	return handler.invitationCommand(context, handler.service.DeclineInvitation)
}

func (handler *Families) CancelInvitation(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	invitationID, err := routeID(context, "invitation_id")
	if err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	if err := handler.service.CancelInvitation(context.Request().Context(), identity.UserID, familyID, invitationID, httpmiddleware.CurrentRequestID(context)); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Families) familyCommand(echoContext *echo.Context, command func(context.Context, platform.ID, platform.ID, string) error) error {
	familyID, err := routeID(echoContext, "family_id")
	if err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(echoContext)
	if err := command(echoContext.Request().Context(), identity.UserID, familyID, httpmiddleware.CurrentRequestID(echoContext)); err != nil {
		return domainError(err)
	}
	return echoContext.NoContent(http.StatusNoContent)
}

func (handler *Families) invitationCommand(echoContext *echo.Context, command func(context.Context, platform.ID, platform.ID, string) error) error {
	invitationID, err := routeID(echoContext, "invitation_id")
	if err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(echoContext)
	if err := command(echoContext.Request().Context(), identity.UserID, invitationID, httpmiddleware.CurrentRequestID(echoContext)); err != nil {
		return domainError(err)
	}
	return echoContext.NoContent(http.StatusNoContent)
}

func routeID(context *echo.Context, name string) (platform.ID, error) {
	id, err := parseID(context.Param(name))
	if err != nil {
		return 0, apierrors.New(http.StatusNotFound, apierrors.CodeNotFound, "The requested resource was not found")
	}
	return id, nil
}

func parseID(raw string) (platform.ID, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, errors.New("invalid ID")
	}
	return platform.ID(value), nil
}
