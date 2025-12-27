package handlers

import (
	"backend-final-project-ponpes/models"
	"backend-final-project-ponpes/services"
	"backend-final-project-ponpes/utils"
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	// Implementation for creating a user will go here

	var reqUser models.UserDTO

	if err := c.ShouldBindJSON(&reqUser); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	message, err := h.userService.CreateUser(reqUser)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, message, nil)

	// loginResponse, err := h.authService.Login(loginReq)
	// if err != nil {
	// 	utils.SendError(c, http.StatusUnauthorized, "Invalid username or password")
	// 	return
	// }
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var reqUser models.UserDTO

	if err := c.ShouldBindJSON(&reqUser); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	message, err := h.userService.UpdatedUser(reqUser)
	fmt.Printf("message: %s", message)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, message, nil)
}
