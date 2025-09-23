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
		product.GET("", productController.GetProducts)            // Get all products
		product.GET("/:id", productController.GetProduct)         // Get product by ID
		product.GET("/search", productController.GetProductBySku) // Get product by SKU

		// Admin product management routes (admin, manager roles)
		productAdmin := product.Group("")
		productAdmin.Use(middleware.RequireProductManagementRoles())
		{
			productAdmin.POST("", productController.CreateProduct)       // Create new product
			productAdmin.PUT("/:id", productController.UpdateProduct)    // Update product by ID
			productAdmin.DELETE("/:id", productController.RemoveProduct) // Delete product by ID
		}
	}
}
