package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupOrderMobileRoutes configures mobile order-related routes for pickers
func SetupOrderMobileRoutes(api *gin.RouterGroup, cfg *config.Config, orderMobileController *controllers.OrderMobileController) {
	// Mobile order routes (authenticated pickers only)
	mobileOrders := api.Group("/mobile/orders")
	mobileOrders.Use(middleware.AuthMiddleware(cfg))
	{
		// Get all orders ready to pick
		mobileOrders.GET("", orderMobileController.GetOrders)
		
		// Pick an order (change status to "picking process")
		mobileOrders.PUT("/:id/pick", orderMobileController.PickingOrder)
		
		// Get order details with product info (location, barcode)
		mobileOrders.GET("/:id", orderMobileController.GetOrder)
		
		// Complete picking process (change status to "picking complete")
		mobileOrders.PUT("/:id/complete", orderMobileController.CompletePickingOrder)
	}
}