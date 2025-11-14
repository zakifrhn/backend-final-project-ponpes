package routes

import (
    "backend-final-project-ponpes/handlers"
    "backend-final-project-ponpes/middleware"
    "github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler) {
    public := router.Group("/api/v1")
    {
        public.POST("/login", authHandler.Login)
    }

    protected := router.Group("/api/v1")
    protected.Use(middleware.AuthMiddleware())
    {
        // Profile routes
        protected.GET("/profile", authHandler.GetProfile)

        santri := protected.Group("/santri")
        santri.Use(middleware.RoleMiddleware("santri", "ustad", "admin"))
        {
            // Add santri routes here
        }

        // Ustad routes - only for ustad and admin
        ustad := protected.Group("/ustad")
        ustad.Use(middleware.RoleMiddleware("ustad", "admin"))
        {
            // Add ustad routes here
        }

        // Admin only routes
        admin := protected.Group("/admin")
        admin.Use(middleware.RoleMiddleware("admin"))
        {
            // Add admin routes here
        }

        // Orang tua routes
        orangTua := protected.Group("/orang-tua")
        orangTua.Use(middleware.RoleMiddleware("orang_tua", "admin"))
        {
            // Add orang tua routes here
        }
    }
}