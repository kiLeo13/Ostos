package families

import (
	"context"
	"time"

	"ostos/backend/internal/platform"
)

type Repository interface {
	WithinTransaction(context.Context, func(Repository) error) error
	LockUser(context.Context, platform.ID) error
	CountUserFamilies(context.Context, platform.ID) (int64, error)
	CreateFamily(context.Context, *Family) error
	ListFamilies(context.Context, platform.ID) ([]Summary, error)
	FindFamily(context.Context, platform.ID) (*Family, error)
	FindFamilyForUpdate(context.Context, platform.ID) (*Family, error)
	FindRole(context.Context, platform.ID, platform.ID) (Role, error)
	ListMembers(context.Context, platform.ID) ([]MemberView, error)
	UpdateFamily(context.Context, platform.ID, string, time.Time) error
	DeleteFamily(context.Context, platform.ID) error
	FindMemberForUpdate(context.Context, platform.ID, platform.ID) (*Member, error)
	FindMemberByUserForUpdate(context.Context, platform.ID, platform.ID) (*Member, error)
	CreateMember(context.Context, *Member) error
	UpdateMemberRole(context.Context, platform.ID, Role, time.Time) error
	DeleteMember(context.Context, platform.ID) error
	ResolveInviteeByUsername(context.Context, string) (platform.ID, error)
	UserExists(context.Context, platform.ID) (bool, error)
	CreateInvitation(context.Context, *Invitation) error
	ListReceivedInvitations(context.Context, platform.ID, time.Time) ([]InvitationView, error)
	ListFamilyInvitations(context.Context, platform.ID, time.Time) ([]SentInvitationView, error)
	FindInvitationForUpdate(context.Context, platform.ID) (*Invitation, error)
	FindFamilyInvitationForUpdate(context.Context, platform.ID, platform.ID) (*Invitation, error)
	DeleteInvitation(context.Context, platform.ID) error
	UpdateOwner(context.Context, platform.ID, platform.ID, time.Time) error
	CreateAuditEvent(context.Context, *AuditEvent) error
}
