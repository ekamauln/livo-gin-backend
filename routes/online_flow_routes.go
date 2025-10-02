package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupOnlineFlowRoutes configures online flow-related routes
func SetupOnlineFlowRoutes(r *gin.RouterGroup, cfg *config.Config, onlineFlowController *controllers.OnlineFlowController) {
	// Online flow routes (authenticated)
	onlineFlow := r.Group("/online_flows")
	onlineFlow.Use(middleware.AuthMiddleware(cfg))
	{
		// Public online flow routes
		onlineFlow.GET("", onlineFlowController.GetOnlineFlows)          // Get all online flows (with optional search)
		onlineFlow.GET("/:tracking", onlineFlowController.GetOnlineFlow) // Get online flow by tracking
	}
}
