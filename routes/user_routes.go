package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes configures user-related routes
func SetupUserRoutes(api *gin.RouterGroup, cfg *config.Config, userController *controllers.UserController) {
	// User routes (authenticated)
	user := api.Group("/user")
	user.Use(middleware.AuthMiddleware(cfg))
	{
		// Public user routes
		user.GET("/profile", userController.GetProfile)    // Get user profile
		user.PUT("/profile", userController.UpdateProfile) // Update user profile
	}
}
