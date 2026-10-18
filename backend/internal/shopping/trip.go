package shopping

import (
	"time"

	"ostos/backend/internal/platform"
)

type Trip struct {
	ID              platform.ID `json:"id" gorm:"primaryKey"`
	FamilyID        platform.ID `json:"family_id" gorm:"column:family_id"`
	SupermarketName *string     `json:"supermarket_name" gorm:"column:supermarket_name"`
	CompletedByID   platform.ID `json:"completed_by_id" gorm:"column:completed_by_id"`
	CompletedAt     time.Time   `json:"completed_at" gorm:"column:completed_at"`
	CreatedAt       time.Time   `json:"created_at" gorm:"column:created_at"`
}

func (Trip) TableName() string {
	return "shopping_trips"
}

type TripDetail struct {
	Trip
	Items []TripItem `json:"items"`
}

type UncheckedAction string

const (
	ActionKeep    UncheckedAction = "keep"
	ActionDiscard UncheckedAction = "discard"
)

func (action UncheckedAction) Valid() bool {
	return action == ActionKeep || action == ActionDiscard
}

type ItemActionOverride struct {
	ItemID platform.ID
	Action UncheckedAction
}

type CompleteListInput struct {
	SupermarketName        *string
	DefaultUncheckedAction UncheckedAction
	UncheckedItemOverrides []ItemActionOverride
}
