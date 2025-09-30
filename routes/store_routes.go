package routes

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	"livo-gin-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupStoreRoutes configures store-related routes
func SetupStoreRoutes(api *gin.RouterGroup, cfg *config.Config, storeController *controllers.StoreController) {
	// Store routes (authenticated)
	store := api.Group("/stores")
	store.Use(middleware.AuthMiddleware(cfg))
	{
		// Public store routes
		store.GET("", storeController.GetStores)          // Get all stores (with optional search)
		store.GET("/:id", storeController.GetStore)       // Get store by ID
		store.POST("", storeController.CreateStore)       // Create new store
		store.PUT("/:id", storeController.UpdateStore)    // Update store by ID
		store.DELETE("/:id", storeController.RemoveStore) // Delete store by ID
	}
}
