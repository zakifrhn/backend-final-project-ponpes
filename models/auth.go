package models

import "github.com/golang-jwt/jwt/v4"

type JWTClaims struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

type AuthResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Token   string `json:"token,omitempty"`
    User    interface{} `json:"user,omitempty"`
}