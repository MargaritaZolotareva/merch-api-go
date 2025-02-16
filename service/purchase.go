package service

import (
	"fmt"
	"gorm.io/gorm"
	"merch-api/model"
)

type PurchaseService interface {
	PurchaseMerch(db *gorm.DB, username string, itemName string) (string, error)
}

type PurchaseServiceImpl struct{}

func NewPurchaseService() *PurchaseServiceImpl {
	return &PurchaseServiceImpl{}
}

func (s *PurchaseServiceImpl) PurchaseMerch(db *gorm.DB, username string, itemName string) (string, error) {
	var merch model.Merch
	var employee model.Employee

	if err := db.Where("name = ?", itemName).First(&merch).Error; err != nil {
		return "", fmt.Errorf("товар %s не найден", itemName)
	}

	tx := db.Begin()
	if tx.Error != nil {
		return "", tx.Error
	}

	if err := tx.Where("username = ?", username).First(&employee).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("пользователь %s не найден", username)
	}

	if employee.Balance < merch.Price {
		tx.Rollback()
		return "", fmt.Errorf("недостаточно монет для покупки товара %s", itemName)
	}

	newBalance := employee.Balance - merch.Price
	employee.Balance = newBalance
	if err := tx.Save(&employee).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("не удалось обновить баланс сотрудника")
	}

	purchase := model.Purchase{
		EmployeeID: employee.ID,
		MerchID:    merch.ID,
	}
	if err := tx.Create(&purchase).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("не удалось сохранить покупку")
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("не удалось зафиксировать транзакцию: %v", err)
	}

	return "Покупка успешна", nil
}
