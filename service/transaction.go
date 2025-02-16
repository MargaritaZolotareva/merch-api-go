package service

import (
	"fmt"
	"gorm.io/gorm"
	"merch-api/model"
)

type TransactionService interface {
	SendCoins(db *gorm.DB, fromUsername, toUsername string, amount int) (string, error)
}

type TransactionServiceImpl struct{}

func NewTransactionService() *TransactionServiceImpl {
	return &TransactionServiceImpl{}
}

func (s *TransactionServiceImpl) SendCoins(db *gorm.DB, fromUsername, toUsername string, amount int) (string, error) {
	var fromEmployee model.Employee
	var toEmployee model.Employee

	if err := db.Where("username = ?", fromUsername).First(&fromEmployee).Error; err != nil {
		return "", fmt.Errorf("пользователь %s не найден", fromUsername)
	}

	if err := db.Where("username = ?", toUsername).First(&toEmployee).Error; err != nil {
		return "", fmt.Errorf("пользователь %s не найден", toUsername)
	}

	if fromEmployee.Balance < amount {
		return "", fmt.Errorf("недостаточно монет на балансе пользователя %s", fromUsername)
	}

	tx := db.Begin()
	if tx.Error != nil {
		return "", tx.Error
	}

	newFromBalance := fromEmployee.Balance - amount
	if err := tx.Model(&fromEmployee).Update("balance", newFromBalance).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("не удалось обновить баланс отправителя")
	}

	newToBalance := toEmployee.Balance + amount
	if err := tx.Model(&toEmployee).Update("balance", newToBalance).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("не удалось обновить баланс получателя")
	}

	transaction := model.Transaction{
		SenderID:   fromEmployee.ID,
		ReceiverID: toEmployee.ID,
		Amount:     amount,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("не удалось создать запись о переводе")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("не удалось зафиксировать транзакцию: %v", err)
	}

	return fmt.Sprintf("Перевод успешен! Кол-во: %d монет пользователю %s. Новый баланс: отправитель %d, получатель %d", amount, toUsername, newFromBalance, newToBalance), nil
}
