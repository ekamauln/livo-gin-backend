package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupReturnRoutes configures return-related routes
func SetupReturnRoutes(api *gin.RouterGroup, cfg *config.Config, returnController *controllers.ReturnController) {
	// Return routes (authenticated)
	returns := api.Group("/returns")
	returns.Use(middleware.AuthMiddleware(cfg))
	{
		// Public return routes
		returns.POST("", returnController.CreateBaseReturn)
		returns.GET("", returnController.GetReturns)                  // Get all returns (with optional search and date filtering)
		returns.GET("/:id", returnController.GetReturn)               // Get return by ID
		returns.PUT("/:id/data", returnController.UpdateDataReturn)   // Update partial data return for normal admins
		returns.PUT("/:id/admin", returnController.UpdateAdminReturn) // Update full return for return admins

	}
}
