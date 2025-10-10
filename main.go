// Package main provides the main entry point for the Livotech Backend Service
// @title Livotech Backend Service API
// @version 1.0
// @description A comprehensive user management backend service with JWT authentication and role-based access control
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host 192.168.31.136:8000
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the 'Bearer ' prefix, e.g. 'Bearer abcdef12345'.

package main

import (
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	_ "livo-gin-backend/docs" // This is required for Swagger
	"livo-gin-backend/migrations"
	"livo-gin-backend/routes"
	"log"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	config.ConnectDatabase(cfg)

	// Run migrations
	db := config.GetDB()
	migrations.AutoMigrate(db)

	// Initialize controllers
	authController := controllers.NewAuthController(db, cfg)
	userController := controllers.NewUserController(db)
	adminController := controllers.NewAdminController(db)
	productController := controllers.NewProductController(db)
	orderController := controllers.NewOrderController(db)
	orderMobileController := controllers.NewOrderMobileController(db)
	boxController := controllers.NewBoxController(db)
	expeditionController := controllers.NewExpeditionController(db)
	channelController := controllers.NewChannelController(db)
	storeController := controllers.NewStoreController(db)
	mbOnlineController := controllers.NewMbOnlineController(db)
	mbRibbonController := controllers.NewMbRibbonController(db)
	qcRibbonController := controllers.NewQcRibbonController(db)
	qcOnlineController := controllers.NewQcOnlineController(db)
	pcOnlineController := controllers.NewPcOnlineController(db)
	outboundController := controllers.NewOutboundController(db)
	onlineFlowController := controllers.NewOnlineFlowController(db)
	ribbonFlowController := controllers.NewRibbonFlowController(db)
	returnController := controllers.NewReturnController(db)
	returnMobileController := controllers.NewReturnMobileController(db)
	complainController := controllers.NewComplainController(db)

	// Setup routes
	router := routes.SetupRoutes(cfg, authController, userController, adminController, productController, orderController, orderMobileController, boxController, expeditionController, channelController, storeController, mbOnlineController, mbRibbonController, qcRibbonController, qcOnlineController, pcOnlineController, outboundController, onlineFlowController, ribbonFlowController, returnController, returnMobileController, complainController)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Health check available at http://192.168.31.136:%s/health", cfg.Port)
	log.Printf("RapiDoc API documentation available at http://192.168.31.136:%s/docs", cfg.Port)
	log.Printf("Legacy Swagger UI still available at http://192.168.31.136:%s/swagger/index.html", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
