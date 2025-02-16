package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	handler2 "merch-api/handler"
	middleware2 "merch-api/middleware"
	service2 "merch-api/service"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	purchaseService := service2.NewPurchaseService()
	purchaseHandler := handler2.NewPurchaseHandler(purchaseService)

	transactionService := service2.NewTransactionService()
	transactionHandler := handler2.NewTransactionHandler(transactionService)

	userInfoService := service2.NewUserInfoService()
	userInfoHandler := handler2.NewUserInfoHandler(userInfoService)

	authService := service2.NewAuthService()
	authHandler := handler2.NewAuthHandler(authService)

	r.Use(middleware2.DatabaseMiddleware(db))
	r.POST("/api/auth", authHandler.Authenticate)

	r.Use(middleware2.JWTMiddleware())
	r.GET("/api/buy/:item", purchaseHandler.BuyItem)
	r.POST("/api/sendCoin", transactionHandler.SendCoin)
	r.GET("/api/info", userInfoHandler.InfoHandler)

	return r
}
