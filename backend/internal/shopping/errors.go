package shopping

import "errors"

var (
	ErrFamilyNotFound   = errors.New("family not found")
	ErrListItemNotFound = errors.New("shopping list item not found")
	ErrTripNotFound     = errors.New("shopping trip not found")
	ErrForbidden        = errors.New("shopping action is forbidden")
	ErrInvalidQuantity  = errors.New("quantity must be positive")
	ErrInvalidUnit      = errors.New("invalid quantity unit")
	ErrInvalidCursor    = errors.New("invalid pagination cursor")
	ErrInvalidAction    = errors.New("invalid unchecked action")
)
