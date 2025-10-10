package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupReturnMobileRoutes configures return-mobile-related routes
func SetupReturnMobileRoutes(api *gin.RouterGroup, cfg *config.Config, returnMobileController *controllers.ReturnMobileController) {
	// ReturnMobile routes (authenticated)
	returnMobile := api.Group("/mobile/returns")
	returnMobile.Use(middleware.AuthMiddleware(cfg))
	{
		// Public return-mobile routes
		returnMobile.GET("", returnMobileController.GetReturnMobiles)
		returnMobile.GET("/:id", returnMobileController.GetReturnMobile)
		returnMobile.POST("", returnMobileController.CreateReturnMobile)
	}
}
