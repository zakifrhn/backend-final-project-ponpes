package routes

import (
	"backend-final-project-ponpes/handlers"
	"backend-final-project-ponpes/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	invoiceHandler *handlers.InvoiceHandler,
	paymentHandler *handlers.PaymentHandler) {

	public := router.Group("/api/v1")
	{
		public.POST("/login", authHandler.Login)

		// Webhook untuk payment gateway (tanpa auth) - REAL TRANSACTION
		public.POST("/payment/webhook", paymentHandler.HandlePaymentWebhook)
	}

	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		// Profile routes
		protected.GET("/profile", authHandler.GetProfile)

		// ==================== PAYMENT ROUTES (REAL) ====================
		paymentRoutes := protected.Group("/payment")
		{
			// Create real payment transaction
			paymentRoutes.POST("/create", paymentHandler.CreatePayment)

			// Create payment for specific invoice (with payment method)
			paymentRoutes.POST("/invoice/:id/pay", paymentHandler.CreatePaymentForInvoice)

			// Validate payment amount
			paymentRoutes.GET("/validate/:id_invoice", paymentHandler.ValidatePayment)

			// Update transaction status
			paymentRoutes.PUT("/transactions/:id/status", paymentHandler.UpdateTransactionStatus)
		}

		// ==================== SANTRI ROUTES ====================
		santri := protected.Group("/santri")
		santri.Use(middleware.RoleMiddleware("santri", "ustad", "admin"))
		{
			// Invoice routes untuk santri
			santri.GET("/invoices", invoiceHandler.GetMyInvoices)
			santri.GET("/invoices/:id", invoiceHandler.GetInvoiceDetail)
			santri.POST("/invoices/:id/pay", paymentHandler.CreatePaymentForInvoice)

			// Transaction history untuk santri (REAL TRANSACTIONS)
			santri.GET("/transactions", paymentHandler.GetMyTransactionHistory)
		}

		// ==================== USTAD ROUTES ====================
		ustad := protected.Group("/ustad")
		ustad.Use(middleware.RoleMiddleware("ustad", "admin"))
		{
			// Invoice management untuk ustad
			ustad.GET("/invoices", invoiceHandler.GetAllInvoices)
			ustad.GET("/invoices/santri/:id_santri", invoiceHandler.GetInvoicesBySantri)
			ustad.POST("/invoices", invoiceHandler.CreateInvoice)
			ustad.PUT("/invoices/:id", invoiceHandler.UpdateInvoice)
			ustad.DELETE("/invoices/:id", invoiceHandler.DeleteInvoice)
			ustad.PUT("/invoices/:id/status", invoiceHandler.UpdateInvoiceStatus)

			// Transaction management untuk ustad (REAL TRANSACTIONS)
			ustad.GET("/transactions", paymentHandler.GetAllTransactions)
			ustad.GET("/transactions/santri/:id_santri", paymentHandler.GetTransactionsBySantri)
			ustad.PUT("/transactions/:id/status", paymentHandler.UpdateTransactionStatus)

			// Payment monitoring
			ustad.GET("/payment/config", paymentHandler.GetPaymentConfig)
		}

		// ==================== ADMIN ROUTES ====================
		admin := protected.Group("/admin")
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			// User management
			admin.POST("/users", userHandler.CreateUser)
			admin.PUT("/users", userHandler.UpdateUser)

			// Invoice automation untuk admin
			admin.POST("/invoices/generate-monthly", invoiceHandler.GenerateMonthlyInvoices)

			// Payment gateway configuration untuk admin (REAL CONFIG)
			admin.GET("/payment/config", paymentHandler.GetPaymentConfig)
			admin.PUT("/payment/config", paymentHandler.UpdatePaymentConfig)
		}

		// ==================== ORANG TUA ROUTES ====================
		orangTua := protected.Group("/orang-tua")
		orangTua.Use(middleware.RoleMiddleware("orang_tua", "admin"))
		{
			// Orang tua bisa melihat invoice anak-anaknya
			orangTua.GET("/invoices", invoiceHandler.GetMyChildrenInvoices)
			orangTua.GET("/invoices/:id", invoiceHandler.GetInvoiceDetail)
			orangTua.POST("/invoices/:id/pay", paymentHandler.CreatePaymentForInvoice)

			// Orang tua bisa melihat transaksi anak-anaknya (REAL TRANSACTIONS)
			orangTua.GET("/transactions", paymentHandler.GetMyChildrenTransactions)
		}
	}

	// ==================== HEALTH CHECK & MONITORING ====================
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "Ponpes Payment Gateway",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	router.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "Backend Ponpes Al-Aziziyah",
			"version": "1.0.0",
			"payment": gin.H{
				"gateway":   "Midtrans",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		})
	})
}
