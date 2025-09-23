package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes configures admin-related routes
func SetupAdminRoutes(api *gin.RouterGroup, cfg *config.Config, adminController *controllers.AdminController) {
	// Admin routes (authenticated + role-based)
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg))
	{
		// User management (superadmin)
		users := admin.Group("/users")
		users.Use(middleware.RequireManagementRoles())
		{
			users.GET("", adminController.GetUsers)                        // List all users
			users.GET("/:id", adminController.GetUser)                     // Get user by ID
			users.PUT("/:id/status", adminController.UpdateUserStatus)     // Update user status (active/inactive)
			users.PUT("/:id/password", adminController.UpdateUserPassword) // Update user password
			users.PUT("/:id/profile", adminController.UpdateUserProfile)   // Update user profile
			users.POST("", adminController.CreateUser)                     // Create new user
			users.DELETE("/:id", adminController.DeleteUser)               // Delete user
		}

		// Role assignment (superadmin, admin)
		roleAssignment := admin.Group("/users/:id/roles") // Assign or remove roles to/from a user
		roleAssignment.Use(middleware.RequireRoleAssignmentRoles())
		{
			roleAssignment.POST("", adminController.AssignRole)   // Assign role to user
			roleAssignment.DELETE("", adminController.RemoveRole) // Remove role from user
		}

		// Roles management (all authenticated users can view)
		admin.GET("/roles", adminController.GetRoles) // List all roles
	}
}
