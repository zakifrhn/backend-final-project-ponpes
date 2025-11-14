package main

import (
    "backend-final-project-ponpes/config"
    "backend-final-project-ponpes/handlers"
    "backend-final-project-ponpes/repositories"
    "backend-final-project-ponpes/routes"
    "backend-final-project-ponpes/services"
    "github.com/gin-gonic/gin"
)

func main() {
    db := config.ConnectDB()
    defer db.Close()
    userRepo := repositories.NewUserRepository(db)
    authService := services.NewAuthService(userRepo)
    authHandler := handlers.NewAuthHandler(authService)
    router := gin.Default()
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
    routes.SetupRoutes(router, authHandler)
    router.Run(":8080")
}