package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupMbOnlineRoutes configures mb-online-related routes
func SetupMbOnlineRoutes(api *gin.RouterGroup, cfg *config.Config, mbOnlineController *controllers.MbOnlineController) {
	// Mb-Online routes (authenticated)
	mbOnline := api.Group("/mb-onlines")
	mbOnline.Use(middleware.AuthMiddleware(cfg))
	{
		// Public mb-online routes
		mbOnline.POST("", mbOnlineController.CreateMbOnline) // Create new mb-online entry
		mbOnline.GET("", mbOnlineController.GetMbOnlines)    // Get all mb-onlines (with optional search and date filtering)
		mbOnline.GET("/:id", mbOnlineController.GetMbOnline) // Get mb-online by ID
	}
}
