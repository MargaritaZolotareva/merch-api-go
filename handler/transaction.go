package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"merch-api/service"
	"net/http"
)

type TransactionHandler struct {
	service service.TransactionService
}

func NewTransactionHandler(svc service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		service: svc,
	}
}

func (h *TransactionHandler) SendCoin(c *gin.Context) {
	fromUsername, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "Неавторизован"})
		return
	}

	fromUsernameString, ok := fromUsername.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "Некорректный username"})
		return
	}

	var input struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Неверный запрос"})
		return
	}
	if input.ToUser == fromUsernameString {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Нельзя отправить монеты самому себе"})
		return
	}

	db, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Database not available"})
		return
	}

	gdb, ok := db.(*gorm.DB)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Invalid database connection"})
		return
	}

	message, err := h.service.SendCoins(gdb, fromUsernameString, input.ToUser, input.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}
