package model

import "time"

type Merch struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"unique;not null"`
	Price int    `gorm:"not null"`
}

func (Merch) TableName() string {
	return "merch"
}

type Purchase struct {
	ID         uint      `gorm:"primaryKey"`
	EmployeeID uint      `gorm:"not null"`
	MerchID    uint      `gorm:"not null"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (Purchase) TableName() string {
	return "purchase"
}

type Transaction struct {
	ID         uint      `gorm:"primaryKey"`
	SenderID   uint      `gorm:"not null"`
	ReceiverID uint      `gorm:"not null"`
	Amount     int       `gorm:"not null"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (Transaction) TableName() string {
	return "transaction"
}

type Employee struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Balance  int    `gorm:"not null"`
}

func (Employee) TableName() string {
	return "employee"
}
