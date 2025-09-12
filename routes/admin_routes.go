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
		// User management (superadmin, admin, manager)
		users := admin.Group("/users")
		users.Use(middleware.RequireManagementRoles())
		{
			users.GET("", adminController.GetUsers)
			users.GET("/:id", adminController.GetUser)
			users.PUT("/:id/status", adminController.UpdateUserStatus)
		}

		// Role assignment (superadmin, admin, manager, supervisor)
		roleAssignment := admin.Group("/users/:id/roles")
		roleAssignment.Use(middleware.RequireRoleAssignmentRoles())
		{
			roleAssignment.POST("", adminController.AssignRole)
			roleAssignment.DELETE("", adminController.RemoveRole)
		}

		// Roles management (all authenticated users can view)
		admin.GET("/roles", adminController.GetRoles)
	}
}
