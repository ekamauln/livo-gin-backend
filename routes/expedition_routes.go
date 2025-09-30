package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupExpeditionRoutes configures expedition-related routes
func SetupExpeditionRoutes(api *gin.RouterGroup, cfg *config.Config, expeditionController *controllers.ExpeditionController) {
	// Expedition routes (authenticated)
	expedition := api.Group("/expeditions")
	expedition.Use(middleware.AuthMiddleware(cfg))
	{
		// Public expedition routes
		expedition.GET("", expeditionController.GetExpeditions)          // Get all expeditions (with optional search)
		expedition.GET("/:id", expeditionController.GetExpedition)       // Get expedition by ID
		expedition.POST("", expeditionController.CreateExpedition)       // Create new expedition
		expedition.PUT("/:id", expeditionController.UpdateExpedition)    // Update expedition by ID
		expedition.DELETE("/:id", expeditionController.RemoveExpedition) // Delete expedition by ID
	}
}
