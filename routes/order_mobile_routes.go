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
	mobileOrders.Use(middleware.RequirePickerRole())
	{
		// Mobile order routes - GetOrders now includes search functionality via query parameter
		mobileOrders.GET("", orderMobileController.GetOrders)                         // Get all orders ready to pick (with optional search)
		mobileOrders.GET("/my-picking", orderMobileController.GetMyPickingOrders)     // Get my ongoing picking orders (status: "picking process")
		mobileOrders.PUT("/:id/pick", orderMobileController.PickingOrder)             // Pick an order (change status to "picking process")
		mobileOrders.GET("/:id", orderMobileController.GetOrder)                      // Get order details with product info (location, barcode)
		mobileOrders.PUT("/:id/complete", orderMobileController.CompletePickingOrder) // Complete picking process (change status to "picking complete")
		mobileOrders.PUT("/:id/cancel", orderMobileController.CancelPickingOrder)     // Cancel picking process (change status back to "ready to pick")
	}
}
