package families

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"ostos/backend/internal/platform"
)

const (
	maxFamiliesPerUser = 10
	invitationTTL      = 7 * 24 * time.Hour
)

type Service struct {
	repository Repository
	ids        platform.IDGenerator
	clock      platform.Clock
}

func NewService(repository Repository, ids platform.IDGenerator, clock platform.Clock) (*Service, error) {
	if repository == nil || ids == nil || clock == nil {
		return nil, errors.New("all family service dependencies are required")
	}
	return &Service{repository: repository, ids: ids, clock: clock}, nil
}

func (service *Service) List(ctx context.Context, actorID platform.ID) ([]Summary, error) {
	return service.repository.ListFamilies(ctx, actorID)
}

func (service *Service) Create(ctx context.Context, actorID platform.ID, name, requestID string) (Family, error) {
	now := service.clock.Now().UTC()
	id, err := service.ids.Next(ctx)
	if err != nil {
		return Family{}, err
	}
	family := Family{ID: id, Name: name, OwnerID: actorID, CreatedAt: now, UpdatedAt: now}
	err = service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if err := repository.LockUser(ctx, actorID); err != nil {
			return err
		}
		count, err := repository.CountUserFamilies(ctx, actorID)
		if err != nil {
			return err
		}
		if count >= maxFamiliesPerUser {
			return ErrFamilyLimit
		}
		if err := repository.CreateFamily(ctx, &family); err != nil {
			return err
		}
		return service.audit(ctx, repository, actorID, family.ID, "family.created", "family", &family.ID, requestID, nil, now)
	})
	return family, err
}

func (service *Service) Get(ctx context.Context, actorID, familyID platform.ID) (Family, error) {
	if _, err := service.authorize(ctx, service.repository, actorID, familyID, ActionView); err != nil {
		return Family{}, err
	}
	family, err := service.repository.FindFamily(ctx, familyID)
	if err != nil {
		return Family{}, err
	}
	return *family, nil
}

func (service *Service) Update(ctx context.Context, actorID, familyID platform.ID, name, requestID string) (Family, error) {
	now := service.clock.Now().UTC()
	if _, err := service.authorize(ctx, service.repository, actorID, familyID, ActionEditSettings); err != nil {
		return Family{}, err
	}
	if err := service.repository.UpdateFamily(ctx, familyID, name, now); err != nil {
		return Family{}, err
	}
	family, err := service.repository.FindFamily(ctx, familyID)
	if err != nil {
		return Family{}, err
	}
	family.Name, family.UpdatedAt = name, now
	return *family, nil
}

func (service *Service) Delete(ctx context.Context, actorID, familyID platform.ID, requestID string) error {
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if _, err := service.authorize(ctx, repository, actorID, familyID, ActionDelete); err != nil {
			return err
		}
		if err := service.audit(ctx, repository, actorID, familyID, "family.deleted", "family", &familyID, requestID, nil, now); err != nil {
			return err
		}
		return repository.DeleteFamily(ctx, familyID)
	})
}

func (service *Service) Members(ctx context.Context, actorID, familyID platform.ID) ([]MemberView, error) {
	if _, err := service.authorize(ctx, service.repository, actorID, familyID, ActionView); err != nil {
		return nil, err
	}
	return service.repository.ListMembers(ctx, familyID)
}

func (service *Service) ChangeRole(ctx context.Context, actorID, familyID, memberID platform.ID, role Role, requestID string) error {
	if role != RoleManager && role != RoleMember {
		return ErrInvalidRole
	}
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if _, err := service.authorize(ctx, repository, actorID, familyID, ActionManageManager); err != nil {
			return err
		}
		member, err := repository.FindMemberForUpdate(ctx, familyID, memberID)
		if err != nil {
			return err
		}
		if err := repository.UpdateMemberRole(ctx, member.ID, role, now); err != nil {
			return err
		}
		return service.audit(ctx, repository, actorID, familyID, "member.role_changed", "member", &member.ID, requestID, map[string]any{"role": role}, now)
	})
}

func (service *Service) RemoveMember(ctx context.Context, actorID, familyID, memberID platform.ID, requestID string) error {
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		actorRole, err := service.authorize(ctx, repository, actorID, familyID, ActionRemoveMember)
		if err != nil {
			return err
		}
		member, err := repository.FindMemberForUpdate(ctx, familyID, memberID)
		if err != nil {
			return err
		}
		if member.Role == RoleManager && actorRole != RoleOwner {
			return ErrForbidden
		}
		if err := repository.DeleteMember(ctx, member.ID); err != nil {
			return err
		}
		return service.audit(ctx, repository, actorID, familyID, "member.removed", "member", &member.ID, requestID, nil, now)
	})
}

func (service *Service) Leave(ctx context.Context, actorID, familyID platform.ID, requestID string) error {
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		role, err := repository.FindRole(ctx, familyID, actorID)
		if err != nil {
			return err
		}
		if role == RoleOwner {
			return ErrOwnerCannotLeave
		}
		member, err := repository.FindMemberByUserForUpdate(ctx, familyID, actorID)
		if err != nil {
			return err
		}
		if err := repository.DeleteMember(ctx, member.ID); err != nil {
			return err
		}
		return service.audit(ctx, repository, actorID, familyID, "member.left", "member", &member.ID, requestID, nil, now)
	})
}

type InviteTarget struct {
	Username *string
	UserID   *platform.ID
}

func (service *Service) Invite(ctx context.Context, actorID, familyID platform.ID, target InviteTarget, requestID string) error {
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if _, err := service.authorize(ctx, repository, actorID, familyID, ActionInviteMember); err != nil {
			return err
		}
		var invitedUserID platform.ID
		var err error
		if target.Username != nil {
			invitedUserID, err = repository.ResolveInviteeByUsername(ctx, *target.Username)
			if err != nil && !errors.Is(err, ErrInviteeNotEligible) {
				return err
			}
		} else if target.UserID != nil {
			invitedUserID = *target.UserID
			var exists bool
			exists, err = repository.UserExists(ctx, invitedUserID)
			if err != nil {
				return err
			}
			if !exists {
				err = ErrInviteeNotEligible
			}
		} else {
			return ErrInviteeNotEligible
		}
		if err != nil || invitedUserID == actorID {
			return ErrInviteeNotEligible
		}
		if err := repository.LockUser(ctx, invitedUserID); err != nil {
			if errors.Is(err, ErrInviteeNotEligible) {
				return ErrInviteeNotEligible
			}
			return err
		}
		count, err := repository.CountUserFamilies(ctx, invitedUserID)
		if err != nil {
			return err
		}
		if count >= maxFamiliesPerUser {
			return ErrInviteeNotEligible
		}
		if _, roleErr := repository.FindRole(ctx, familyID, invitedUserID); roleErr == nil {
			return ErrInviteeNotEligible
		}
		id, err := service.ids.Next(ctx)
		if err != nil {
			return err
		}
		invitation := &Invitation{ID: id, FamilyID: familyID, InvitedUserID: invitedUserID,
			InvitedByUserID: actorID, ExpiresAt: now.Add(invitationTTL), CreatedAt: now}
		if err := repository.CreateInvitation(ctx, invitation); err != nil {
			return err
		}
		return service.audit(ctx, repository, actorID, familyID, "invitation.created", "invitation", &id, requestID, nil, now)
	})
}

func (service *Service) ReceivedInvitations(ctx context.Context, actorID platform.ID) ([]InvitationView, error) {
	return service.repository.ListReceivedInvitations(ctx, actorID, service.clock.Now().UTC())
}

func (service *Service) FamilyInvitations(ctx context.Context, actorID, familyID platform.ID) ([]SentInvitationView, error) {
	if _, err := service.authorize(ctx, service.repository, actorID, familyID, ActionInviteMember); err != nil {
		return nil, err
	}
	return service.repository.ListFamilyInvitations(ctx, familyID, service.clock.Now().UTC())
}

func (service *Service) AcceptInvitation(ctx context.Context, actorID, invitationID platform.ID, requestID string) error {
	now := service.clock.Now().UTC()
	var expired bool
	err := service.repository.WithinTransaction(ctx, func(repository Repository) error {
		invitation, err := repository.FindInvitationForUpdate(ctx, invitationID)
		if err != nil || invitation.InvitedUserID != actorID {
			return ErrInvitationNotFound
		}
		if !invitation.ExpiresAt.After(now) {
			expired = true
			return repository.DeleteInvitation(ctx, invitation.ID)
		}
		if err := repository.LockUser(ctx, actorID); err != nil {
			return err
		}
		count, err := repository.CountUserFamilies(ctx, actorID)
		if err != nil {
			return err
		}
		if count >= maxFamiliesPerUser {
			return ErrFamilyLimit
		}
		memberID, err := service.ids.Next(ctx)
		if err != nil {
			return err
		}
		member := &Member{ID: memberID, FamilyID: invitation.FamilyID, UserID: actorID,
			Role: RoleMember, CreatedAt: now, UpdatedAt: now}
		if err := repository.CreateMember(ctx, member); err != nil {
			return err
		}
		if err := repository.DeleteInvitation(ctx, invitation.ID); err != nil {
			return err
		}
		return service.audit(ctx, repository, actorID, invitation.FamilyID, "invitation.accepted", "invitation", &invitation.ID, requestID, nil, now)
	})
	if err == nil && expired {
		return ErrInvitationNotFound
	}
	return err
}

func (service *Service) DeclineInvitation(ctx context.Context, actorID, invitationID platform.ID, requestID string) error {
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		invitation, err := repository.FindInvitationForUpdate(ctx, invitationID)
		if err != nil || invitation.InvitedUserID != actorID {
			return ErrInvitationNotFound
		}
		if err := repository.DeleteInvitation(ctx, invitation.ID); err != nil {
			return err
		}
		return service.audit(ctx, repository, actorID, invitation.FamilyID, "invitation.declined", "invitation", &invitation.ID, requestID, nil, now)
	})
}

func (service *Service) CancelInvitation(ctx context.Context, actorID, familyID, invitationID platform.ID, requestID string) error {
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if _, err := service.authorize(ctx, repository, actorID, familyID, ActionInviteMember); err != nil {
			return err
		}
		invitation, err := repository.FindFamilyInvitationForUpdate(ctx, familyID, invitationID)
		if err != nil {
			return err
		}
		if err := repository.DeleteInvitation(ctx, invitation.ID); err != nil {
			return err
		}
		return service.audit(ctx, repository, actorID, familyID, "invitation.cancelled", "invitation", &invitation.ID, requestID, nil, now)
	})
}

func (service *Service) TransferOwnership(ctx context.Context, actorID, familyID, newOwnerID platform.ID, requestID string) error {
	now := service.clock.Now().UTC()
	return service.repository.WithinTransaction(ctx, func(repository Repository) error {
		family, err := repository.FindFamilyForUpdate(ctx, familyID)
		if err != nil || family.OwnerID != actorID {
			return ErrForbidden
		}
		newOwnerMembership, err := repository.FindMemberByUserForUpdate(ctx, familyID, newOwnerID)
		if err != nil {
			return ErrInvalidOwner
		}
		oldOwnerMembershipID, err := service.ids.Next(ctx)
		if err != nil {
			return err
		}
		if err := repository.DeleteMember(ctx, newOwnerMembership.ID); err != nil {
			return err
		}
		oldOwnerMembership := &Member{ID: oldOwnerMembershipID, FamilyID: familyID, UserID: actorID,
			Role: RoleManager, CreatedAt: now, UpdatedAt: now}
		if err := repository.CreateMember(ctx, oldOwnerMembership); err != nil {
			return err
		}
		if err := repository.UpdateOwner(ctx, familyID, newOwnerID, now); err != nil {
			return err
		}
		return service.audit(ctx, repository, actorID, familyID, "family.ownership_transferred", "user", &newOwnerID, requestID, nil, now)
	})
}

func (service *Service) authorize(ctx context.Context, repository Repository, actorID, familyID platform.ID, action Action) (Role, error) {
	role, err := repository.FindRole(ctx, familyID, actorID)
	if err != nil {
		return "", ErrFamilyNotFound
	}
	if !Allows(role, action) {
		return role, ErrForbidden
	}
	return role, nil
}

func (service *Service) audit(ctx context.Context, repository Repository, actorID, familyID platform.ID, action, targetType string, targetID *platform.ID, requestID string, metadata map[string]any, now time.Time) error {
	id, err := service.ids.Next(ctx)
	if err != nil {
		return err
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	event := &AuditEvent{ID: id, ActorUserID: &actorID, FamilyID: &familyID, ActionCode: action,
		TargetType: targetType, TargetID: targetID, Metadata: raw, RequestID: requestID, CreatedAt: now}
	return repository.CreateAuditEvent(ctx, event)
}
