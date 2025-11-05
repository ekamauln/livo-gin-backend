package routes

import (
	"livo-gin-backend/controllers"

	"github.com/gin-gonic/gin"
)

// SetupStoreMobileRoutes configures store-mobile-related routes
func SetupStoreMobileRoutes(api *gin.RouterGroup, storeMobileController *controllers.StoreMobileController) {
	// StoreMobile routes (public - no authentication required)
	storeMobile := api.Group("/mobile/stores")
	{
		storeMobile.GET("", storeMobileController.GetStoreMobiles)
	}
}
