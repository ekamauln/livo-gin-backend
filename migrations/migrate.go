package migrations

import (
	"livo-gin-backend/models"
	"log"

	"gorm.io/gorm"
)

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) {
	// Run migrations
	err := db.AutoMigrate(
		&models.Role{},
		&models.User{},
		&models.UserRole{},
		&models.Product{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed successfully")

	// Seed default roles
	seedDefaultRoles(db)
}

// seedDefaultRoles creates default roles if they don't exist
func seedDefaultRoles(db *gorm.DB) {
	roles := []models.Role{
		{Name: "superadmin", Description: "Super Administrator with full system access"},
		{Name: "admin", Description: "Administrator with high-level management access"},
		{Name: "manager", Description: "Manager with user and operational management access"},
		{Name: "supervisor", Description: "Supervisor with team management access"},
		{Name: "picker", Description: "Picker with basic operational access"},
	}

	for _, role := range roles {
		var existingRole models.Role
		if err := db.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			// Role doesn't exist, create it
			if err := db.Create(&role).Error; err != nil {
				log.Printf("Failed to create role %s: %v", role.Name, err)
			} else {
				log.Printf("Created role: %s", role.Name)
			}
		}
	}
}
