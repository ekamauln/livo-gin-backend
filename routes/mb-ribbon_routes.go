package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupMbRibbonRoutes configures mb-ribbon-related routes
func SetupMbRibbonRoutes(api *gin.RouterGroup, cfg *config.Config, mbRibbonController *controllers.MbRibbonController) {
	// Mb-Ribbon routes (authenticated)
	mbRibbon := api.Group("/mb-ribbons")
	mbRibbon.Use(middleware.AuthMiddleware(cfg))
	{
		// Public mb-ribbon routes
		mbRibbon.POST("", mbRibbonController.CreateMbRibbon)
		mbRibbon.GET("", mbRibbonController.GetMbRibbons)    // Get all mb-ribbons (with optional search and date filtering)
		mbRibbon.GET("/:id", mbRibbonController.GetMbRibbon) // Get mb-ribbon by ID
	}
}
