package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRibbonFlowRoutes configures ribbon flow-related routes
func SetupRibbonFlowRoutes(r *gin.RouterGroup, cfg *config.Config, ribbonFlowController *controllers.RibbonFlowController) {
	// Ribbon flow routes (authenticated)
	ribbonFlow := r.Group("/ribbon-flow")
	ribbonFlow.Use(middleware.AuthMiddleware(cfg))
	{
		// Public ribbon flow routes
		ribbonFlow.GET("", ribbonFlowController.GetRibbonFlows)          // Get all ribbon flows (with optional search)
		ribbonFlow.GET("/:tracking", ribbonFlowController.GetRibbonFlow) // Get ribbon flow by tracking
	}
}
