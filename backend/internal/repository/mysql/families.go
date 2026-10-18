package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"ostos/backend/internal/accounts"
	"ostos/backend/internal/families"
	"ostos/backend/internal/platform"
)

type FamilyRepository struct{ database *gorm.DB }

func NewFamilyRepository(database *gorm.DB) *FamilyRepository {
	return &FamilyRepository{database: database}
}

func (repository *FamilyRepository) WithinTransaction(ctx context.Context, operation func(families.Repository) error) error {
	return WithinTransaction(ctx, repository.database, nil, func(transaction *gorm.DB) error {
		return operation(&FamilyRepository{database: transaction})
	})
}

func (repository *FamilyRepository) LockUser(ctx context.Context, userID platform.ID) error {
	var user accounts.User
	err := repository.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", userID).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return families.ErrInviteeNotEligible
	}
	return err
}

func (repository *FamilyRepository) CountUserFamilies(ctx context.Context, userID platform.ID) (int64, error) {
	var count int64
	err := repository.database.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM (
			SELECT id FROM families WHERE owner_id = ? AND deleted_at IS NULL
			UNION ALL
			SELECT fm.family_id FROM family_members fm JOIN families f ON f.id = fm.family_id
			WHERE fm.user_id = ? AND f.deleted_at IS NULL
		) AS participation`, userID, userID).Scan(&count).Error
	return count, err
}

func (repository *FamilyRepository) CreateFamily(ctx context.Context, family *families.Family) error {
	return repository.database.WithContext(ctx).Create(family).Error
}

func (repository *FamilyRepository) ListFamilies(ctx context.Context, userID platform.ID) ([]families.Summary, error) {
	var summaries []families.Summary
	err := repository.database.WithContext(ctx).Raw(`
		SELECT f.id, f.name, 'owner' AS role, f.created_at, f.updated_at
		FROM families f WHERE f.owner_id = ? AND f.deleted_at IS NULL
		UNION ALL
		SELECT f.id, f.name, fm.role, f.created_at, f.updated_at
		FROM family_members fm JOIN families f ON f.id = fm.family_id
		WHERE fm.user_id = ? AND f.deleted_at IS NULL
		ORDER BY name, id`, userID, userID).Scan(&summaries).Error
	return summaries, err
}

func (repository *FamilyRepository) FindFamily(ctx context.Context, familyID platform.ID) (*families.Family, error) {
	var family families.Family
	err := repository.database.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", familyID).Take(&family).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, families.ErrFamilyNotFound
	}
	return &family, err
}

func (repository *FamilyRepository) FindFamilyForUpdate(ctx context.Context, familyID platform.ID) (*families.Family, error) {
	var family families.Family
	err := repository.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", familyID).Take(&family).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, families.ErrFamilyNotFound
	}
	return &family, err
}

func (repository *FamilyRepository) FindRole(ctx context.Context, familyID, userID platform.ID) (families.Role, error) {
	var family families.Family
	err := repository.database.WithContext(ctx).Select("id", "owner_id").Where("id = ? AND deleted_at IS NULL", familyID).Take(&family).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", families.ErrFamilyNotFound
	}
	if err != nil {
		return "", err
	}
	if family.OwnerID == userID {
		return families.RoleOwner, nil
	}
	var member families.Member
	err = repository.database.WithContext(ctx).Select("role").Where("family_id = ? AND user_id = ?", familyID, userID).Take(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", families.ErrFamilyNotFound
	}
	return member.Role, err
}

func (repository *FamilyRepository) ListMembers(ctx context.Context, familyID platform.ID) ([]families.MemberView, error) {
	var members []families.MemberView
	err := repository.database.WithContext(ctx).Raw(`
		SELECT NULL AS id, u.id AS user_id, u.username, u.display_name, 'owner' AS role, f.created_at AS joined_at
		FROM families f JOIN users u ON u.id = f.owner_id WHERE f.id = ? AND f.deleted_at IS NULL
		UNION ALL
		SELECT fm.id, u.id AS user_id, u.username, u.display_name, fm.role, fm.created_at AS joined_at
		FROM family_members fm JOIN users u ON u.id = fm.user_id JOIN families f ON f.id = fm.family_id
		WHERE fm.family_id = ? AND f.deleted_at IS NULL
		ORDER BY CASE role WHEN 'owner' THEN 0 WHEN 'manager' THEN 1 ELSE 2 END, username`, familyID, familyID).Scan(&members).Error
	return members, err
}

func (repository *FamilyRepository) UpdateFamily(ctx context.Context, familyID platform.ID, name string, now time.Time) error {
	result := repository.database.WithContext(ctx).Model(&families.Family{}).Where("id = ? AND deleted_at IS NULL", familyID).Updates(map[string]any{"name": name, "updated_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return families.ErrFamilyNotFound
	}
	return nil
}

func (repository *FamilyRepository) DeleteFamily(ctx context.Context, familyID platform.ID) error {
	result := repository.database.WithContext(ctx).Where("id = ?", familyID).Delete(&families.Family{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return families.ErrFamilyNotFound
	}
	return nil
}

func (repository *FamilyRepository) FindMemberForUpdate(ctx context.Context, familyID, memberID platform.ID) (*families.Member, error) {
	var member families.Member
	err := repository.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("family_id = ? AND id = ?", familyID, memberID).Take(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, families.ErrMemberNotFound
	}
	return &member, err
}

func (repository *FamilyRepository) FindMemberByUserForUpdate(ctx context.Context, familyID, userID platform.ID) (*families.Member, error) {
	var member families.Member
	err := repository.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("family_id = ? AND user_id = ?", familyID, userID).Take(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, families.ErrMemberNotFound
	}
	return &member, err
}

func (repository *FamilyRepository) CreateMember(ctx context.Context, member *families.Member) error {
	return repository.database.WithContext(ctx).Create(member).Error
}

func (repository *FamilyRepository) UpdateMemberRole(ctx context.Context, memberID platform.ID, role families.Role, now time.Time) error {
	result := repository.database.WithContext(ctx).Model(&families.Member{}).Where("id = ?", memberID).Updates(map[string]any{"role": role, "updated_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return families.ErrMemberNotFound
	}
	return nil
}

func (repository *FamilyRepository) DeleteMember(ctx context.Context, memberID platform.ID) error {
	result := repository.database.WithContext(ctx).Where("id = ?", memberID).Delete(&families.Member{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return families.ErrMemberNotFound
	}
	return nil
}

func (repository *FamilyRepository) ResolveInviteeByUsername(ctx context.Context, username string) (platform.ID, error) {
	var user accounts.User
	err := repository.database.WithContext(ctx).Select("id").Where("username = ? AND deleted_at IS NULL", username).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, families.ErrInviteeNotEligible
	}
	return user.ID, err
}

func (repository *FamilyRepository) UserExists(ctx context.Context, userID platform.ID) (bool, error) {
	var count int64
	err := repository.database.WithContext(ctx).Model(&accounts.User{}).Where("id = ? AND deleted_at IS NULL", userID).Count(&count).Error
	return count == 1, err
}

func (repository *FamilyRepository) CreateInvitation(ctx context.Context, invitation *families.Invitation) error {
	err := repository.database.WithContext(ctx).Create(invitation).Error
	if duplicateKey(err, "uq_family_invitations_family_user") {
		return families.ErrInviteeNotEligible
	}
	return err
}

func (repository *FamilyRepository) ListReceivedInvitations(ctx context.Context, userID platform.ID, now time.Time) ([]families.InvitationView, error) {
	var invitations []families.InvitationView
	err := repository.database.WithContext(ctx).Table("family_invitations AS i").Select("i.id, i.family_id, f.name AS family_name, i.expires_at, i.created_at").
		Joins("JOIN families AS f ON f.id = i.family_id AND f.deleted_at IS NULL").Where("i.invited_user_id = ? AND i.expires_at > ?", userID, now).
		Order("i.created_at DESC").Scan(&invitations).Error
	return invitations, err
}

func (repository *FamilyRepository) ListFamilyInvitations(ctx context.Context, familyID platform.ID, now time.Time) ([]families.SentInvitationView, error) {
	var invitations []families.SentInvitationView
	err := repository.database.WithContext(ctx).Table("family_invitations AS i").Select("i.id, u.id AS user_id, u.username, u.display_name, i.expires_at, i.created_at").
		Joins("JOIN users AS u ON u.id = i.invited_user_id").Where("i.family_id = ? AND i.expires_at > ?", familyID, now).
		Order("i.created_at DESC").Scan(&invitations).Error
	return invitations, err
}

func (repository *FamilyRepository) FindInvitationForUpdate(ctx context.Context, invitationID platform.ID) (*families.Invitation, error) {
	var invitation families.Invitation
	err := repository.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", invitationID).Take(&invitation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, families.ErrInvitationNotFound
	}
	return &invitation, err
}

func (repository *FamilyRepository) FindFamilyInvitationForUpdate(ctx context.Context, familyID, invitationID platform.ID) (*families.Invitation, error) {
	var invitation families.Invitation
	err := repository.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("family_id = ? AND id = ?", familyID, invitationID).Take(&invitation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, families.ErrInvitationNotFound
	}
	return &invitation, err
}

func (repository *FamilyRepository) DeleteInvitation(ctx context.Context, invitationID platform.ID) error {
	result := repository.database.WithContext(ctx).Where("id = ?", invitationID).Delete(&families.Invitation{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return families.ErrInvitationNotFound
	}
	return nil
}

func (repository *FamilyRepository) UpdateOwner(ctx context.Context, familyID, newOwnerID platform.ID, now time.Time) error {
	result := repository.database.WithContext(ctx).Model(&families.Family{}).Where("id = ?", familyID).Updates(map[string]any{"owner_id": newOwnerID, "updated_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return families.ErrFamilyNotFound
	}
	return nil
}

func (repository *FamilyRepository) CreateAuditEvent(ctx context.Context, event *families.AuditEvent) error {
	return repository.database.WithContext(ctx).Create(event).Error
}
