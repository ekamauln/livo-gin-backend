package models

import (
	"time"

	"gorm.io/gorm"
)

// Role represents system roles
type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"unique;not null" json:"name" example:"admin"`
	Description string         `json:"description" example:"Administrator role"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// GetRoleHierarchy returns role hierarchy levels
func GetRoleHierarchy() map[string]int {
	return map[string]int{
		"superadmin": 5,
		"admin":      4,
		"manager":    3,
		"supervisor": 2,
		"picker":     1,
	}
}

// CanManageRole checks if a role can manage another role
func (r *Role) CanManageRole(targetRole string) bool {
	hierarchy := GetRoleHierarchy()
	currentLevel, exists := hierarchy[r.Name]
	if !exists {
		return false
	}

	targetLevel, exists := hierarchy[targetRole]
	if !exists {
		return false
	}

	return currentLevel > targetLevel
}
