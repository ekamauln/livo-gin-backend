package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupQcOnlineRoutes configures qc-online-related routes
func SetupQcOnlineRoutes(api *gin.RouterGroup, cfg *config.Config, qcOnlineController *controllers.QcOnlineController) {
	// Qc-Online routes (authenticated)
	qcOnline := api.Group("/qc-onlines")
	qcOnline.Use(middleware.AuthMiddleware(cfg))
	{
		// Public qc-online routes
		qcOnline.GET("", qcOnlineController.GetQcOnlines)    // Get all qc-onlines (with optional search and date filtering)
		qcOnline.GET("/:id", qcOnlineController.GetQcOnline) // Get qc-online by ID
		qcOnline.POST("", qcOnlineController.CreateQcOnline) // Create new qc-online
	}
}
