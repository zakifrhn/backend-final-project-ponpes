package main

import (
	"backend-final-project-ponpes/config"
	"backend-final-project-ponpes/handlers"
	"backend-final-project-ponpes/repositories"
	"backend-final-project-ponpes/routes"
	"backend-final-project-ponpes/services"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Koneksi database
	db := config.ConnectDB()
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("✅ Database connected")

	// ==================== PAYMENT CONFIGURATION ====================
	// Gunakan config.GetPaymentConfig() untuk mendapatkan konfigurasi yang benar
	paymentConfig := config.GetPaymentConfig()

	log.Printf("=== PAYMENT GATEWAY CONFIGURATION ===")
	log.Printf("Environment: %s", map[bool]string{true: "PRODUCTION", false: "SANDBOX"}[paymentConfig.IsProduction])
	log.Printf("Server Key: %s...", maskString(paymentConfig.ServerKey, 8))
	log.Printf("Client Key: %s...", maskString(paymentConfig.ClientKey, 8))
	log.Printf("=====================================")

	// ==================== INITIALIZE REPOSITORIES ====================
	userRepo := repositories.NewUserRepository(db)
	invoiceRepo := repositories.NewInvoiceRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	santriRepo := repositories.NewSantriRepository(db)

	// ==================== INITIALIZE SERVICES ====================
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	invoiceService := services.NewInvoiceService(invoiceRepo, santriRepo, db)

	// Payment Service TANPA santriRepo (sesuai dengan perubahan)
	paymentService := services.NewPaymentService(
		db,
		invoiceRepo,
		transactionRepo,
		santriRepo,
		paymentConfig.ServerKey,
		paymentConfig.ClientKey,
		paymentConfig.IsProduction,
	)

	// ==================== INITIALIZE HANDLERS ====================
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService)
	paymentHandler := handlers.NewPaymentHandler(paymentService)

	// ==================== SETUP ROUTER ====================
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// ==================== SETUP ROUTES ====================
	routes.SetupRoutes(router, authHandler, userHandler, invoiceHandler, paymentHandler)

	// ==================== START SERVER ====================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server started on port %s", port)
	log.Printf("📱 Payment Gateway: %s", map[bool]string{true: "PRODUCTION", false: "SANDBOX"}[paymentConfig.IsProduction])
	log.Printf("🌐 Webhook URL: %s/api/v1/payment/webhook", os.Getenv("APP_URL"))

	router.Run(":" + port)
}

// Helper function untuk mask sensitive data
func maskString(str string, visibleChars int) string {
	if len(str) <= visibleChars*2 {
		return "***"
	}
	return str[:visibleChars] + "..." + str[len(str)-visibleChars:]
}
