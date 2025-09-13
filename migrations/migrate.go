package migrations

import (
	"livo-gin-backend/models"
	"livo-gin-backend/utils"
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

	// Seed first superadmin user
	seedSuperadminUser(db)
}

// seedDefaultRoles creates default roles if they don't exist
func seedDefaultRoles(db *gorm.DB) {
	roles := []models.Role{
		{Name: "superadmin", Description: "Super Administrator with full system access"},
		{Name: "admin", Description: "Administrator with high-level management access"},
		{Name: "finance", Description: "Finance role with financial management access"},
		{Name: "picker", Description: "Picker with basic operational access"},
		{Name: "outbound", Description: "Outbound role with shipping management access"},
		{Name: "qc-ribbon", Description: "Quality Control for Ribbon products"},
		{Name: "qc-online", Description: "Quality Control for Online products"},
		{Name: "mb-ribbon", Description: "Product Checker for Ribbon products"},
		{Name: "mb-online", Description: "Product Checker for Online products"},
		{Name: "packing", Description: "Packing role with packaging management access"},
		{Name: "guest", Description: "Guest with limited access"},
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

// seedSuperadminUser creates the first superadmin user if it doesn't exist
func seedSuperadminUser(db *gorm.DB) {
	// Check if superadmin user already exists
	var existingUser models.User
	if err := db.Where("username = ?", "saya").First(&existingUser).Error; err == nil {
		log.Println("Superadmin user already exists")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword("gajahku")
	if err != nil {
		log.Printf("Failed to hash superadmin password: %v", err)
		return
	}

	// Create superadmin user
	user := models.User{
		Username: "saya",
		Email:    "eka.mauln@gmail.com",
		Password: hashedPassword,
		FullName: "Saya",
		IsActive: true,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Printf("Failed to create superadmin user: %v", err)
		return
	}

	// Find superadmin role
	var superadminRole models.Role
	if err := db.Where("name = ?", "superadmin").First(&superadminRole).Error; err != nil {
		log.Printf("Superadmin role not found: %v", err)
		return
	}

	// Assign superadmin role
	userRole := models.UserRole{
		UserID:     user.ID,
		RoleID:     superadminRole.ID,
		AssignedBy: user.ID, // Self-assigned for the first superadmin
	}

	if err := db.Create(&userRole).Error; err != nil {
		log.Printf("Failed to assign superadmin role: %v", err)
		return
	}

	log.Println("Superadmin user created successfully")
}
