package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupPickOrderRoutes configures pick order-related routes
func SetupPickOrderRoutes(api *gin.RouterGroup, cfg *config.Config, pickOrderController *controllers.PickOrderController) {
	// Pick Order routes (authenticated)
	pickOrders := api.Group("/pick-orders")
	pickOrders.Use(middleware.AuthMiddleware(cfg))
	{
		// Public pick order routes
		pickOrders.GET("", pickOrderController.GetPickOrders) // Get all pick orders (with optional search and date filtering)
		pickOrders.GET("/:id", pickOrderController.GetPickOrder)
	}
}
