package mysql

import "gorm.io/gorm"

// FamilyRepository is the future MySQL adapter for the families capability.
type FamilyRepository struct{ database *gorm.DB }

func NewFamilyRepository(database *gorm.DB) *FamilyRepository {
	return &FamilyRepository{database: database}
}
