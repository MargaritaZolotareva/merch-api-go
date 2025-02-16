package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"merch-api/service"
	"net/http"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: svc,
	}
}

func (h *AuthHandler) Authenticate(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "invalid input"})
		return
	}

	token, err := h.service.AuthenticateUser(db, input.Username, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"errors": "invalid input"})
		case errors.Is(err, service.ErrPasswordMismatch):
			c.JSON(http.StatusUnauthorized, gin.H{"errors": "invalid password"})
		case errors.Is(err, service.ErrFailedToCreateUser):
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "failed to find or create employee"})
		case errors.Is(err, service.ErrFailedToGenerateToken):
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "failed to generate token"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
