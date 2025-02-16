package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"merch-api/service"
	"net/http"
)

type PurchaseHandler struct {
	service service.PurchaseService
}

func NewPurchaseHandler(svc service.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{
		service: svc,
	}
}

func (h *PurchaseHandler) BuyItem(c *gin.Context) {
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

	itemName := c.Param("item")

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

	message, err := h.service.PurchaseMerch(gdb, usernameString, itemName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}
