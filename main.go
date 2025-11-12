package main

import (
	"fmt"
	"livo-gin-backend/config"
	"livo-gin-backend/controllers"
	_ "livo-gin-backend/docs" // This is required for Swagger
	"livo-gin-backend/migrations"
	"livo-gin-backend/routes"
	"log"
)

// @title Livotech Backend Service API
// @version 1.0
// @description A comprehensive user management backend service with JWT authentication and role-based access control
// @contact.name API Support
// @contact.email support@livotech.com
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	log.Println("🚀 Starting Livotech Backend Service...")

	// Load configuration
	log.Println("📝 Loading configuration...")
	cfg := config.LoadConfig()
	log.Println("✓ Configuration loaded successfully")

	// Connect to database with retry logic
	log.Println("🔌 Connecting to database...")
	config.ConnectDatabase(cfg)

	// Run migrations
	log.Println("🔄 Running database migrations...")
	db := config.GetDB()
	migrations.AutoMigrate(db) // No error handling needed, it's handled inside the function

	// Initialize controllers
	log.Println("🎮 Initializing controllers...")
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
	reportController := controllers.NewReportController(db)
	channelMobileController := controllers.NewChannelMobileController(db)
	storeMobileController := controllers.NewStoreMobileController(db)
	pickOrderController := controllers.NewPickOrderController(db)
	log.Println("✓ Controllers initialized successfully")

	// Setup routes
	log.Println("🛣️  Setting up routes...")
	router := routes.SetupRoutes(cfg, authController, userController, adminController, productController, orderController, orderMobileController, boxController, expeditionController, channelController, storeController, mbOnlineController, mbRibbonController, qcRibbonController, qcOnlineController, pcOnlineController, outboundController, onlineFlowController, ribbonFlowController, returnController, returnMobileController, complainController, reportController, channelMobileController, storeMobileController, pickOrderController)
	log.Println("✓ Routes configured successfully")

	// Build API URL from config
	apiURL := fmt.Sprintf("http://%s:%s", cfg.APIHost, cfg.Port)

	// Start server
	log.Println("════════════════════════════════════════════════════════════")
	log.Printf("✓ Server ready on port %s", cfg.Port)
	log.Printf("📊 Health check: %s/health", apiURL)
	log.Printf("📚 API documentation: %s/docs", apiURL)
	log.Printf("📖 Swagger UI: %s/swagger/index.html", apiURL)
	log.Println("════════════════════════════════════════════════════════════")

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
