package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"ostos/backend/internal/families"
	"ostos/backend/internal/platform"
	"ostos/backend/internal/shopping"
)

type ShoppingRepository struct {
	database *gorm.DB
}

func NewShoppingRepository(database *gorm.DB) *ShoppingRepository {
	return &ShoppingRepository{database: database}
}

func (repository *ShoppingRepository) WithinTransaction(ctx context.Context, fn func(shopping.Repository) error) error {
	return repository.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewShoppingRepository(tx))
	})
}

func (repository *ShoppingRepository) CheckMembership(ctx context.Context, familyID, userID platform.ID) error {
	var family families.Family
	err := repository.database.WithContext(ctx).Select("id", "owner_id").Where("id = ? AND deleted_at IS NULL", familyID).Take(&family).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shopping.ErrFamilyNotFound
	}
	if err != nil {
		return err
	}
	if family.OwnerID == userID {
		return nil
	}
	var memberCount int64
	err = repository.database.WithContext(ctx).Model(&families.Member{}).Where("family_id = ? AND user_id = ?", familyID, userID).Count(&memberCount).Error
	if err != nil {
		return err
	}
	if memberCount == 0 {
		return shopping.ErrForbidden
	}
	return nil
}

func (repository *ShoppingRepository) LockFamily(ctx context.Context, familyID platform.ID) error {
	var family families.Family
	err := repository.database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", familyID).Take(&family).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shopping.ErrFamilyNotFound
	}
	return err
}

func (repository *ShoppingRepository) GetList(ctx context.Context, familyID platform.ID) ([]shopping.ListItem, error) {
	var items []shopping.ListItem
	err := repository.database.WithContext(ctx).Where("family_id = ?", familyID).Order("position ASC, id ASC").Find(&items).Error
	return items, err
}

func (repository *ShoppingRepository) GetListItem(ctx context.Context, familyID, itemID platform.ID) (*shopping.ListItem, error) {
	var item shopping.ListItem
	err := repository.database.WithContext(ctx).Where("family_id = ? AND id = ?", familyID, itemID).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, shopping.ErrListItemNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (repository *ShoppingRepository) GetMaxPosition(ctx context.Context, familyID platform.ID) (int64, error) {
	var maxPos sql.NullInt64
	err := repository.database.WithContext(ctx).Model(&shopping.ListItem{}).Where("family_id = ?", familyID).Select("COALESCE(MAX(position), 0)").Scan(&maxPos).Error
	if err != nil {
		return 0, err
	}
	return maxPos.Int64, nil
}

func (repository *ShoppingRepository) CreateListItem(ctx context.Context, item *shopping.ListItem) error {
	return repository.database.WithContext(ctx).Create(item).Error
}

func (repository *ShoppingRepository) UpdateListItem(ctx context.Context, item *shopping.ListItem) error {
	result := repository.database.WithContext(ctx).Model(&shopping.ListItem{}).
		Where("family_id = ? AND id = ?", item.FamilyID, item.ID).
		Select("name", "quantity", "unit", "note", "is_checked", "purchased_quantity", "total_cents", "updated_by_id", "updated_at").
		Updates(item)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shopping.ErrListItemNotFound
	}
	return nil
}

func (repository *ShoppingRepository) DeleteListItem(ctx context.Context, familyID, itemID platform.ID) error {
	result := repository.database.WithContext(ctx).Where("family_id = ? AND id = ?", familyID, itemID).Delete(&shopping.ListItem{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shopping.ErrListItemNotFound
	}
	return nil
}

func (repository *ShoppingRepository) ReorderListItems(ctx context.Context, familyID platform.ID, itemIDs []platform.ID) error {
	for index, id := range itemIDs {
		pos := int64((index + 1) * 1000)
		result := repository.database.WithContext(ctx).Model(&shopping.ListItem{}).
			Where("family_id = ? AND id = ?", familyID, id).
			Update("position", pos)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func (repository *ShoppingRepository) LockListItemsForUpdate(ctx context.Context, familyID platform.ID) ([]shopping.ListItem, error) {
	var items []shopping.ListItem
	err := repository.database.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("family_id = ?", familyID).
		Order("id ASC").
		Find(&items).Error
	return items, err
}

func (repository *ShoppingRepository) CreateTrip(ctx context.Context, trip *shopping.Trip, items []shopping.TripItem) error {
	if err := repository.database.WithContext(ctx).Create(trip).Error; err != nil {
		return err
	}
	if len(items) > 0 {
		if err := repository.database.WithContext(ctx).Create(&items).Error; err != nil {
			return err
		}
	}
	return nil
}

func (repository *ShoppingRepository) DeleteListItems(ctx context.Context, familyID platform.ID, itemIDs []platform.ID) error {
	if len(itemIDs) == 0 {
		return nil
	}
	return repository.database.WithContext(ctx).Where("family_id = ? AND id IN ?", familyID, itemIDs).Delete(&shopping.ListItem{}).Error
}

func (repository *ShoppingRepository) ResetUncheckedListItems(ctx context.Context, familyID platform.ID, itemIDs []platform.ID, updatedAt time.Time) error {
	if len(itemIDs) == 0 {
		return nil
	}
	return repository.database.WithContext(ctx).Model(&shopping.ListItem{}).
		Where("family_id = ? AND id IN ?", familyID, itemIDs).
		Updates(map[string]any{
			"is_checked":         false,
			"purchased_quantity": nil,
			"total_cents":        nil,
			"updated_at":         updatedAt,
		}).Error
}

func (repository *ShoppingRepository) ListTrips(ctx context.Context, familyID platform.ID, cursor *platform.ID, limit int) ([]shopping.Trip, *platform.ID, error) {
	var trips []shopping.Trip
	query := repository.database.WithContext(ctx).Where("family_id = ?", familyID)
	if cursor != nil {
		query = query.Where("id < ?", *cursor)
	}
	err := query.Order("id DESC").Limit(limit + 1).Find(&trips).Error
	if err != nil {
		return nil, nil, err
	}
	var nextCursor *platform.ID
	if len(trips) > limit {
		nextCursor = &trips[limit-1].ID
		trips = trips[:limit]
	}
	return trips, nextCursor, nil
}

func (repository *ShoppingRepository) GetTrip(ctx context.Context, familyID, tripID platform.ID) (*shopping.TripDetail, error) {
	var trip shopping.Trip
	err := repository.database.WithContext(ctx).Where("family_id = ? AND id = ?", familyID, tripID).Take(&trip).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, shopping.ErrTripNotFound
	}
	if err != nil {
		return nil, err
	}
	var items []shopping.TripItem
	err = repository.database.WithContext(ctx).Where("trip_id = ?", tripID).Order("id ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []shopping.TripItem{}
	}
	return &shopping.TripDetail{Trip: trip, Items: items}, nil
}
