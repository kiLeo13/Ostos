package request

type ListItemCreateRequest struct {
	Name     string  `json:"name" validate:"required,notblank,min=1,max=160"`
	Quantity string  `json:"quantity" validate:"required,quantity"`
	Unit     string  `json:"unit" validate:"required,oneof=each g kg ml l"`
	Note     *string `json:"note" validate:"omitempty,max=500"`
}

type ListItemUpdateRequest struct {
	Name              *string `json:"name" validate:"omitempty,notblank,min=1,max=160"`
	Quantity          *string `json:"quantity" validate:"omitempty,quantity"`
	Unit              *string `json:"unit" validate:"omitempty,oneof=each g kg ml l"`
	Note              *string `json:"note" validate:"omitempty,max=500"`
	ClearNote         bool    `json:"clear_note"`
	IsChecked         *bool   `json:"is_checked"`
	PurchasedQuantity *string `json:"purchased_quantity" validate:"omitempty,quantity"`
	ClearPurchasedQty bool    `json:"clear_purchased_qty"`
	TotalCents        *int64  `json:"total_cents" validate:"omitempty,gte=0"`
	ClearTotalCents   bool    `json:"clear_total_cents"`
}

type ListItemReorderRequest struct {
	ItemIDs []string `json:"item_ids" validate:"required,min=1,dive,snowflake_id"`
}

type ItemActionOverrideRequest struct {
	ItemID string `json:"item_id" validate:"required,snowflake_id"`
	Action string `json:"action" validate:"required,oneof=keep discard"`
}

type ShoppingListCompleteRequest struct {
	SupermarketName        *string                     `json:"supermarket_name" validate:"omitempty,notblank,max=120"`
	DefaultUncheckedAction string                      `json:"default_unchecked_action" validate:"required,oneof=keep discard"`
	UncheckedItemOverrides []ItemActionOverrideRequest `json:"unchecked_item_overrides" validate:"omitempty,dive"`
}
