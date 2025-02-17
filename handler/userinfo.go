package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"merch-api/service"
	"net/http"
)

type UserInfoHandler struct {
	service service.UserInfoService
}

func NewUserInfoHandler(svc service.UserInfoService) *UserInfoHandler {
	return &UserInfoHandler{
		service: svc,
	}
}

func (h *UserInfoHandler) InfoHandler(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "Неавторизован"})
		return
	}

	usernameString, ok := username.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "Некорректный username"})
		return
	}

	db, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "БД недоступна"})
		return
	}

	gdb, ok := db.(*gorm.DB)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Кривое подключение к БД"})
		return
	}

	userInfo, err := h.service.GetUserInfo(gdb, usernameString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userInfo)
}
