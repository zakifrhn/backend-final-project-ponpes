package utils

import "github.com/gin-gonic/gin"

type Response struct {
	Success int         `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SendSuccess(c *gin.Context, message string, data interface{}) {
	c.JSON(200, Response{
		Success: 200,
		Message: message,
		Data:    data,
	})
}

func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: 400,
		Message: message,
		Error:   message,
	})
}
