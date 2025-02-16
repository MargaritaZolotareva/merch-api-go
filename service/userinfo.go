package service

import (
	"fmt"
	"gorm.io/gorm"
	"merch-api/model"
)

type UserInfo struct {
	Coins     int `json:"coins"`
	Inventory []struct {
		Type     string `json:"type"`
		Quantity int    `json:"quantity"`
	} `json:"inventory"`
	CoinHistory struct {
		Received []struct {
			FromUser string `json:"fromUser" gorm:"column:fromUser"`
			Amount   int    `json:"amount"`
		} `json:"received"`
		Sent []struct {
			ToUser string `json:"toUser" gorm:"column:toUser"`
			Amount int    `json:"amount"`
		} `json:"sent"`
	} `json:"coinHistory"`
}

type UserInfoService interface {
	GetUserInfo(db *gorm.DB, username string) (UserInfo, error)
}

type UserInfoServiceImpl struct{}

func NewUserInfoService() *UserInfoServiceImpl {
	return &UserInfoServiceImpl{}
}

func (s *UserInfoServiceImpl) GetUserInfo(db *gorm.DB, username string) (UserInfo, error) {
	var userInfo UserInfo

	var employee model.Employee
	if err := db.Where("username = ?", username).First(&employee).Error; err != nil {
		return userInfo, fmt.Errorf("пользователь %s не найден", username)
	}
	userInfo.Coins = employee.Balance

	if err := db.Table("purchase").
		Select("merch.name as type, COUNT(purchase.id) as quantity").
		Joins("JOIN merch ON merch.id = purchase.merch_id").
		Where("purchase.employee_id = ?", employee.ID).
		Group("merch.name").
		Scan(&userInfo.Inventory).Error; err != nil {
		return userInfo, fmt.Errorf("не удалось получить инвентарь пользователя: %v", err)
	}

	if err := db.Table("transaction").
		Select("employee.username as \"toUser\", transaction.amount").
		Joins("JOIN employee employee ON transaction.receiver_id = employee.id").
		Where("transaction.sender_id = ?", employee.ID).
		Scan(&userInfo.CoinHistory.Sent).Error; err != nil {
		return userInfo, fmt.Errorf("не удалось получить отправленные транзакции пользователя: %v", err)
	}

	if err := db.Debug().Table("transaction").
		Select("employee.username as \"fromUser\", transaction.amount").
		Joins("JOIN employee employee ON transaction.sender_id = employee.id").
		Where("transaction.receiver_id = ?", employee.ID).
		Scan(&userInfo.CoinHistory.Received).Error; err != nil {
		return userInfo, fmt.Errorf("не удалось получить полученные транзакции пользователя: %v", err)
	}

	return userInfo, nil
}
