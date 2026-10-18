package shopping

import (
	"time"

	"ostos/backend/internal/platform"
)

type TripItem struct {
	ID                platform.ID        `json:"id" gorm:"primaryKey"`
	TripID            platform.ID        `json:"trip_id" gorm:"column:trip_id"`
	SourceItemID      platform.ID        `json:"source_item_id" gorm:"column:source_item_id"`
	Name              string             `json:"name" gorm:"column:name"`
	Quantity          platform.Quantity  `json:"quantity" gorm:"column:quantity"`
	Unit              platform.Unit      `json:"unit" gorm:"column:unit"`
	Note              *string            `json:"note" gorm:"column:note"`
	AddedByID         platform.ID        `json:"added_by_id" gorm:"column:added_by_id"`
	PurchasedQuantity *platform.Quantity `json:"purchased_quantity" gorm:"column:purchased_quantity"`
	TotalCents        *platform.Cents    `json:"total_cents" gorm:"column:total_cents"`
	CreatedAt         time.Time          `json:"created_at" gorm:"column:created_at"`
}

func (TripItem) TableName() string {
	return "shopping_trip_items"
}
