package service

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	service2 "merch-api/service"
	"testing"
)

func TestPurchaseMerch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"merch\" WHERE name = (.+)").
		WithArgs("itemName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
			AddRow(1, "itemName", 100))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("userName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "userName", 200))

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs("userName", "", 100, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO \"purchase\" (.+) VALUES (.+)").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	purchaseService := service2.NewPurchaseService()
	result, err := purchaseService.PurchaseMerch(gdb, "userName", "itemName")
	assert.NoError(t, err)
	assert.Equal(t, "Покупка успешна", result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("не все ожидания выполнены: %v", err)
	}
}

func TestPurchaseMerch_ErrorCreatingPurchase(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"merch\" WHERE name = (.+)").
		WithArgs("itemName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
			AddRow(1, "itemName", 100))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("userName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "userName", 200))

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs("userName", "", 100, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO \"purchase\" (.+) VALUES (.+)").
		WithArgs(1, 1).
		WillReturnError(fmt.Errorf("не удалось сохранить покупку"))

	purchaseService := service2.NewPurchaseService()
	result, err := purchaseService.PurchaseMerch(gdb, "userName", "itemName")

	assert.Error(t, err)
	assert.Equal(t, "не удалось сохранить покупку", err.Error())
	assert.Empty(t, result)
}

func TestPurchaseMerch_ErrorUpdatingEmployeeBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"merch\" WHERE name = (.+)").
		WithArgs("itemName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
			AddRow(1, "itemName", 100))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("userName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "userName", 200))

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs("userName", "", 100, 1).
		WillReturnError(fmt.Errorf("не удалось обновить баланс сотрудника"))

	purchaseService := service2.NewPurchaseService()
	result, err := purchaseService.PurchaseMerch(gdb, "userName", "itemName")

	assert.Error(t, err)
	assert.Equal(t, "не удалось обновить баланс сотрудника", err.Error())
	assert.Empty(t, result)
}

func TestPurchaseMerch_ErrorBeginningTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"merch\" WHERE name = (.+)").
		WithArgs("itemName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
			AddRow(1, "itemName", 100))

	mock.ExpectBegin().WillReturnError(fmt.Errorf("ошибка при начале транзакции"))

	purchaseService := service2.NewPurchaseService()
	result, err := purchaseService.PurchaseMerch(gdb, "userName", "itemName")

	assert.Error(t, err)
	assert.Equal(t, "ошибка при начале транзакции", err.Error())
	assert.Empty(t, result)
}

func TestPurchaseMerch_ErrorCommittingTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"merch\" WHERE name = (.+)").
		WithArgs("itemName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
			AddRow(1, "itemName", 100))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("userName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "userName", 200))

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs("userName", "", 100, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO \"purchase\" (.+) VALUES (.+)").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(fmt.Errorf(""))

	purchaseService := service2.NewPurchaseService()
	result, err := purchaseService.PurchaseMerch(gdb, "userName", "itemName")

	assert.Error(t, err)
	assert.Equal(t, "не удалось зафиксировать транзакцию: ", err.Error())
	assert.Empty(t, result)
}

func TestPurchaseMerch_MerchNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"merch\" WHERE name = (.+)").
		WithArgs("itemName", 1).
		WillReturnError(fmt.Errorf("товар не найден"))

	purchaseService := service2.NewPurchaseService()
	result, err := purchaseService.PurchaseMerch(gdb, "userName", "itemName")
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "товар itemName не найден")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("не все ожидания выполнены: %v", err)
	}
}

func TestPurchaseMerch_UserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"merch\" WHERE name = (.+)").
		WithArgs("itemName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
			AddRow(1, "itemName", 100))
	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("userName", 1).
		WillReturnError(fmt.Errorf("пользователь не найден"))

	purchaseService := service2.NewPurchaseService()
	result, err := purchaseService.PurchaseMerch(gdb, "userName", "itemName")
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "пользователь userName не найден")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("не все ожидания выполнены: %v", err)
	}
}

func TestPurchaseMerch_NotEnoughCoins(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"merch\" WHERE name = (.+)").
		WithArgs("itemName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
			AddRow(1, "itemName", 100))
	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("userName", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "userName", 50))

	purchaseService := service2.NewPurchaseService()
	result, err := purchaseService.PurchaseMerch(gdb, "userName", "itemName")
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "недостаточно монет для покупки товара itemName")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("не все ожидания выполнены: %v", err)
	}
}
