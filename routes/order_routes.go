package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupOrderRoutes configures order-related routes
func SetupOrderRoutes(api *gin.RouterGroup, cfg *config.Config, orderController *controllers.OrderController) {
	// Order routes (authenticated)
	order := api.Group("/orders")
	order.Use(middleware.AuthMiddleware(cfg))
	{
		// Public order routes
		order.GET("", orderController.GetOrders)              // Get all orders
		order.GET("/search", orderController.SearchOrders)    // Search orders
		order.POST("", orderController.CreateOrder)           // Create new order
		order.POST("/bulk", orderController.BulkCreateOrders) // Create multiple orders
	}
}
