package middleware

import (
    "net/http"
    "backend-final-project-ponpes/utils"
    "strings"
    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "message": "Authorization header required",
            })
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "message": "Invalid authorization format",
            })
            c.Abort()
            return
        }

        token := parts[1]
        claims, err := utils.ValidateJWT(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "message": "Invalid or expired token",
            })
            c.Abort()
            return
        }

        c.Set("userID", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)
        c.Next()
    }
}

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("role")
        
        allowed := false
        for _, role := range allowedRoles {
            if userRole == role {
                allowed = true
                break
            }
        }

        if !allowed {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "message": "Insufficient permissions",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}