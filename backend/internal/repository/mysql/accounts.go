package mysql

import "gorm.io/gorm"

// AccountRepository is the future MySQL adapter for the accounts capability.
// Methods are added with account use cases in implementation step 2.
type AccountRepository struct{ database *gorm.DB }

func NewAccountRepository(database *gorm.DB) *AccountRepository {
	return &AccountRepository{database: database}
}
