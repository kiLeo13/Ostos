package shopping

import (
	"context"
	"time"

	"ostos/backend/internal/platform"
)

type Repository interface {
	WithinTransaction(context.Context, func(Repository) error) error
	CheckMembership(ctx context.Context, familyID, userID platform.ID) error
	LockFamily(ctx context.Context, familyID platform.ID) error
	GetList(ctx context.Context, familyID platform.ID) ([]ListItem, error)
	GetListItem(ctx context.Context, familyID, itemID platform.ID) (*ListItem, error)
	GetMaxPosition(ctx context.Context, familyID platform.ID) (int64, error)
	CreateListItem(ctx context.Context, item *ListItem) error
	UpdateListItem(ctx context.Context, item *ListItem) error
	DeleteListItem(ctx context.Context, familyID, itemID platform.ID) error
	ReorderListItems(ctx context.Context, familyID platform.ID, itemIDs []platform.ID) error
	LockListItemsForUpdate(ctx context.Context, familyID platform.ID) ([]ListItem, error)
	CreateTrip(ctx context.Context, trip *Trip, items []TripItem) error
	DeleteListItems(ctx context.Context, familyID platform.ID, itemIDs []platform.ID) error
	ResetUncheckedListItems(ctx context.Context, familyID platform.ID, itemIDs []platform.ID, updatedAt time.Time) error
	ListTrips(ctx context.Context, familyID platform.ID, cursor *platform.ID, limit int) ([]Trip, *platform.ID, error)
	GetTrip(ctx context.Context, familyID, tripID platform.ID) (*TripDetail, error)
}
