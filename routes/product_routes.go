package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupProductRoutes configures product-related routes
func SetupProductRoutes(api *gin.RouterGroup, cfg *config.Config, productController *controllers.ProductController) {
	// Product routes (authenticated)
	product := api.Group("/products")
	product.Use(middleware.AuthMiddleware(cfg))
	{
		// Public product routes
		product.GET("", productController.GetProducts)
		product.GET("/:id", productController.GetProduct)
		product.GET("/search", productController.GetProductBySku)

		// Admin product management routes (admin, manager roles)
		productAdmin := product.Group("")
		productAdmin.Use(middleware.RequireProductManagementRoles())
		{
			productAdmin.POST("", productController.CreateProduct)
			productAdmin.PUT("/:id", productController.UpdateProduct)
			productAdmin.DELETE("/:id", productController.RemoveProduct)
		}
	}
}
