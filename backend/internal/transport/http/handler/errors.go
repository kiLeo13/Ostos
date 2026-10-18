package handler

import (
	"errors"
	"net/http"

	"ostos/backend/internal/accounts"
	"ostos/backend/internal/apierrors"
	"ostos/backend/internal/families"
	"ostos/backend/internal/shopping"
)

func domainError(err error) error {
	switch {
	case errors.Is(err, accounts.ErrEmailTaken):
		return apierrors.New(http.StatusConflict, "EMAIL_ALREADY_REGISTERED", "The email is already registered")
	case errors.Is(err, accounts.ErrUsernameTaken):
		return apierrors.New(http.StatusConflict, "USERNAME_ALREADY_REGISTERED", "The username is already registered")
	case errors.Is(err, accounts.ErrInvalidCredentials):
		return apierrors.New(http.StatusUnauthorized, "INVALID_CREDENTIALS", "The supplied credentials are invalid")
	case errors.Is(err, accounts.ErrEmailNotVerified):
		return apierrors.New(http.StatusForbidden, "EMAIL_NOT_VERIFIED", "Email verification is required")
	case errors.Is(err, accounts.ErrInvalidAccountToken):
		return apierrors.New(http.StatusBadRequest, "INVALID_ACCOUNT_TOKEN", "The account token is invalid or expired")
	case errors.Is(err, accounts.ErrInvalidRefreshToken), errors.Is(err, accounts.ErrRefreshTokenReuse):
		return apierrors.New(http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "The refresh token is invalid")
	case errors.Is(err, accounts.ErrOwnsFamilies):
		return apierrors.New(http.StatusConflict, "OWNED_FAMILIES_REMAIN", "Transfer or delete owned Families first")
	case errors.Is(err, accounts.ErrUserNotFound):
		return apierrors.New(http.StatusNotFound, "USER_NOT_FOUND", "The user was not found")
	case errors.Is(err, families.ErrForbidden):
		return apierrors.New(http.StatusForbidden, "FAMILY_ACTION_FORBIDDEN", "You do not have permission for this action")
	case errors.Is(err, families.ErrFamilyNotFound):
		return apierrors.New(http.StatusNotFound, "FAMILY_NOT_FOUND", "The Family was not found")
	case errors.Is(err, families.ErrMemberNotFound):
		return apierrors.New(http.StatusNotFound, "MEMBER_NOT_FOUND", "The Family member was not found")
	case errors.Is(err, families.ErrInvitationNotFound):
		return apierrors.New(http.StatusNotFound, "INVITATION_NOT_FOUND", "The invitation was not found")
	case errors.Is(err, families.ErrFamilyLimit):
		return apierrors.New(http.StatusConflict, "FAMILY_LIMIT_REACHED", "A user can participate in at most 10 Families")
	case errors.Is(err, families.ErrInviteeNotEligible):
		return apierrors.New(http.StatusBadRequest, "INVITEE_NOT_ELIGIBLE", "The invitee is not eligible")
	case errors.Is(err, families.ErrOwnerCannotLeave):
		return apierrors.New(http.StatusConflict, "OWNER_MUST_TRANSFER", "Transfer ownership before leaving the Family")
	case errors.Is(err, families.ErrInvalidOwner):
		return apierrors.New(http.StatusConflict, "INVALID_NEW_OWNER", "The new Owner must be a current Family member")
	case errors.Is(err, families.ErrInvalidRole):
		return apierrors.InvalidFormBody(map[string]string{"role": "Choose a supported value"})
	case errors.Is(err, shopping.ErrListItemNotFound):
		return apierrors.New(http.StatusNotFound, "LIST_ITEM_NOT_FOUND", "The shopping list item was not found")
	case errors.Is(err, shopping.ErrTripNotFound):
		return apierrors.New(http.StatusNotFound, "TRIP_NOT_FOUND", "The shopping trip was not found")
	case errors.Is(err, shopping.ErrFamilyNotFound):
		return apierrors.New(http.StatusNotFound, "FAMILY_NOT_FOUND", "The Family was not found")
	case errors.Is(err, shopping.ErrForbidden):
		return apierrors.New(http.StatusForbidden, "FORBIDDEN", "You do not have permission for this action")
	case errors.Is(err, shopping.ErrInvalidQuantity):
		return apierrors.InvalidFormBody(map[string]string{"quantity": "Enter a valid positive decimal quantity"})
	case errors.Is(err, shopping.ErrInvalidUnit):
		return apierrors.InvalidFormBody(map[string]string{"unit": "Choose a supported value"})
	case errors.Is(err, shopping.ErrInvalidCursor):
		return apierrors.New(http.StatusBadRequest, "INVALID_CURSOR", "The pagination cursor is invalid")
	case errors.Is(err, shopping.ErrInvalidAction):
		return apierrors.InvalidFormBody(map[string]string{"default_unchecked_action": "Choose keep or discard"})
	default:
		return &apierrors.Error{Status: http.StatusInternalServerError, Code: apierrors.CodeInternal, Message: "An internal error occurred", Cause: err}
	}
}
