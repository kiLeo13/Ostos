package families

import "errors"

var (
	ErrFamilyNotFound     = errors.New("family not found")
	ErrMemberNotFound     = errors.New("family member not found")
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrForbidden          = errors.New("family action is forbidden")
	ErrFamilyLimit        = errors.New("family participation limit reached")
	ErrInviteeNotEligible = errors.New("invitee is not eligible")
	ErrOwnerCannotLeave   = errors.New("owner must transfer ownership before leaving")
	ErrInvalidOwner       = errors.New("new owner must be a current family member")
	ErrInvalidRole        = errors.New("invalid family role")
)
