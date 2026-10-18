package mysql

import "gorm.io/gorm"

// ShoppingRepository is the future MySQL adapter for the shopping capability.
type ShoppingRepository struct{ database *gorm.DB }

func NewShoppingRepository(database *gorm.DB) *ShoppingRepository {
	return &ShoppingRepository{database: database}
}
