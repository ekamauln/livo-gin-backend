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
		order.GET("", orderController.GetOrders)
		order.GET("/search", orderController.SearchOrders)
		order.POST("", orderController.CreateOrder)
		order.POST("/bulk", orderController.BulkCreateOrders)
	}
}
