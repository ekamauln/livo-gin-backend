package routes

import (
	"livo-gin-backend/controllers"

	"github.com/gin-gonic/gin"
)

// SetupReturnMobileRoutes configures return-mobile-related routes
func SetupReturnMobileRoutes(api *gin.RouterGroup, returnMobileController *controllers.ReturnMobileController) {
	// ReturnMobile routes (public - no authentication required)
	returnMobile := api.Group("/mobile/returns")
	{
		returnMobile.GET("", returnMobileController.GetReturnMobiles)
		returnMobile.GET("/:id", returnMobileController.GetReturnMobile)
		returnMobile.POST("", returnMobileController.CreateReturnMobile)
	}
}
