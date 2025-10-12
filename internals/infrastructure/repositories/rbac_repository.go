package repositories

import (
	"gorm.io/gorm"
)

type RBACRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) *RBACRepository {
	return &RBACRepository{db: db}
}

// HasAccess checks if a user has the required permission
func (r *RBACRepository) HasAccess(userID uint, accessName string) bool {
	var count int64

	err := r.db.Table("users u").
		Joins("JOIN role_access ra ON u.role_id = ra.role_id").
		Joins("JOIN accesses a ON ra.access_id = a.id").
		Where("u.id = ? AND a.name = ?", userID, accessName).
		Count(&count).Error

	if err != nil {
		return false
	}

	return count > 0
}
