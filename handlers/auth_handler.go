package handlers

import (
	"backend-final-project-ponpes/models"
	"backend-final-project-ponpes/services"
	"backend-final-project-ponpes/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq models.LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	loginResponse, err := h.authService.Login(loginReq)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}
	responseData := map[string]interface{}{
		"token": loginResponse.Token,
		"user": map[string]interface{}{
			"user_id":    loginResponse.UserID,
			"username":   loginResponse.Username,
			"role":       loginResponse.Role,
			"nama":       loginResponse.Nama,
			"last_login": loginResponse.LastLogin,
		},
	}

	utils.SendSuccess(c, "success login", responseData)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt("userID")
	role := c.GetString("role")
	username := c.GetString("username")

	profileData := map[string]interface{}{
		"user_id":  userID,
		"username": username,
		"role":     role,
	}

	utils.SendSuccess(c, "Profile retrieved successfully", profileData)
}
