package shopping

import (
	"time"

	"ostos/backend/internal/platform"
)

type ListItem struct {
	ID                platform.ID        `json:"id" gorm:"primaryKey"`
	FamilyID          platform.ID        `json:"family_id" gorm:"column:family_id"`
	Name              string             `json:"name" gorm:"column:name"`
	Quantity          platform.Quantity  `json:"quantity" gorm:"column:quantity"`
	Unit              platform.Unit      `json:"unit" gorm:"column:unit"`
	Note              *string            `json:"note" gorm:"column:note"`
	IsChecked         bool               `json:"is_checked" gorm:"column:is_checked"`
	PurchasedQuantity *platform.Quantity `json:"purchased_quantity" gorm:"column:purchased_quantity"`
	TotalCents        *platform.Cents    `json:"total_cents" gorm:"column:total_cents"`
	Position          int64              `json:"position" gorm:"column:position"`
	AddedByID         platform.ID        `json:"added_by_id" gorm:"column:added_by_id"`
	UpdatedByID       platform.ID        `json:"updated_by_id" gorm:"column:updated_by_id"`
	CreatedAt         time.Time          `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time          `json:"updated_at" gorm:"column:updated_at"`
}

func (ListItem) TableName() string {
	return "shopping_list_items"
}

type CreateItemInput struct {
	Name     string
	Quantity platform.Quantity
	Unit     platform.Unit
	Note     *string
}

type UpdateItemInput struct {
	Name              *string
	Quantity          *platform.Quantity
	Unit              *platform.Unit
	Note              *string
	ClearNote         bool
	IsChecked         *bool
	PurchasedQuantity *platform.Quantity
	ClearPurchasedQty bool
	TotalCents        *platform.Cents
	ClearTotalCents   bool
}
