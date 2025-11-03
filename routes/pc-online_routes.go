package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupPcOnlineRoutes configures pc-online-related routes
func SetupPcOnlineRoutes(api *gin.RouterGroup, cfg *config.Config, pcOnlineController *controllers.PcOnlineController) {
	// Pc-Online routes (authenticated)
	pcOnline := api.Group("/pc-onlines")
	pcOnline.Use(middleware.AuthMiddleware(cfg))
	{
		// Public pc-online routes
		pcOnline.GET("", pcOnlineController.GetPcOnlines)            // Get all pc-onlines (with optional search and date filtering)
		pcOnline.GET("/:id", pcOnlineController.GetPcOnline)         // Get pc-online by ID
		pcOnline.POST("", pcOnlineController.CreatePcOnline)         // Create new pc-online
		pcOnline.GET("/chart", pcOnlineController.GetChartPcOnlines) // Get pc-online counts per day for current month
	}
}
