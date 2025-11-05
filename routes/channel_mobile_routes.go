package routes

import (
	"livo-gin-backend/controllers"

	"github.com/gin-gonic/gin"
)

// SetupChannelMobileRoutes configures channel-mobile-related routes
func SetupChannelMobileRoutes(api *gin.RouterGroup, channelMobileController *controllers.ChannelMobileController) {
	// ChannelMobile routes (public - no authentication required)
	channelMobile := api.Group("/mobile/channels")
	{
		channelMobile.GET("", channelMobileController.GetChannelMobiles)
	}
}
