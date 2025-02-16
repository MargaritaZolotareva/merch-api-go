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

func TestSendCoins(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 100))

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user2", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(2, "user2", 50))

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs(90, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs(60, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("INSERT INTO \"transaction\" (.+)").
		WithArgs(1, 2, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	transactionService := service2.NewTransactionService()
	result, err := transactionService.SendCoins(gdb, "user1", "user2", 10)

	assert.NoError(t, err)
	assert.Equal(t, "Перевод успешен! Кол-во: 10 монет пользователю user2. Новый баланс: отправитель 90, получатель 60", result)
}

func TestSendCoins_ErrorCommittingTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 100))

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user2", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(2, "user2", 50))

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs(90, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs(60, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("INSERT INTO \"transaction\" (.+)").
		WithArgs(1, 2, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit().WillReturnError(fmt.Errorf(""))

	transactionService := service2.NewTransactionService()
	result, err := transactionService.SendCoins(gdb, "user1", "user2", 10)

	assert.Error(t, err)
	assert.Equal(t, "не удалось зафиксировать транзакцию: ", err.Error())
	assert.Empty(t, result)
}

func TestSendCoins_ErrorCreatingTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 100))

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user2", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(2, "user2", 50))

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs(90, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs(60, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("INSERT INTO \"transaction\" (.+)").
		WithArgs(1, 2, 10).
		WillReturnError(fmt.Errorf("не удалось создать запись о переводе"))

	transactionService := service2.NewTransactionService()
	result, err := transactionService.SendCoins(gdb, "user1", "user2", 10)

	assert.Error(t, err)
	assert.Equal(t, "не удалось создать запись о переводе", err.Error())
	assert.Empty(t, result)
}

func TestSendCoins_ErrorBeginningTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 100))

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user2", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(2, "user2", 50))

	mock.ExpectBegin().WillReturnError(fmt.Errorf("ошибка при начале транзакции"))

	transactionService := service2.NewTransactionService()
	result, err := transactionService.SendCoins(gdb, "user1", "user2", 10)

	assert.Error(t, err)
	assert.Equal(t, "ошибка при начале транзакции", err.Error())
	assert.Empty(t, result)
}

func TestSendCoins_InsufficientBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {

		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 5))

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user2", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(2, "user2", 50))

	transactionService := service2.NewTransactionService()
	result, err := transactionService.SendCoins(gdb, "user1", "user2", 10)

	assert.Error(t, err)
	assert.Equal(t, "недостаточно монет на балансе пользователя user1", err.Error())
	assert.Empty(t, result)
}

func TestSendCoins_FromUserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnError(fmt.Errorf("пользователь не найден"))

	transactionService := service2.NewTransactionService()
	result, err := transactionService.SendCoins(gdb, "user1", "user2", 10)

	assert.Error(t, err)
	assert.Equal(t, "пользователь user1 не найден", err.Error())
	assert.Empty(t, result)
}

func TestSendCoins_ToUserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 100))

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user2", 1).
		WillReturnError(fmt.Errorf("пользователь не найден"))

	transactionService := service2.NewTransactionService()
	result, err := transactionService.SendCoins(gdb, "user1", "user2", 10)

	assert.Error(t, err)
	assert.Equal(t, "пользователь user2 не найден", err.Error())
	assert.Empty(t, result)
}

func TestSendCoins_UpdateSenderBalanceError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 100))

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user2", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(2, "user2", 50))

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs(90, 1).
		WillReturnError(fmt.Errorf("ошибка при обновлении"))

	transactionService := service2.NewTransactionService()
	result, err := transactionService.SendCoins(gdb, "user1", "user2", 10)

	assert.Error(t, err)
	assert.Equal(t, "не удалось обновить баланс отправителя", err.Error())
	assert.Empty(t, result)
}

func TestSendCoins_UpdateReceiverBalanceError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 100))

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user2", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(2, "user2", 50))

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs(90, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE \"employee\" SET (.+) WHERE (.+)").
		WithArgs(60, 2).
		WillReturnError(fmt.Errorf("ошибка при обновлении"))

	transactionService := service2.NewTransactionService()
	result, err := transactionService.SendCoins(gdb, "user1", "user2", 10)

	assert.Error(t, err)
	assert.Equal(t, "не удалось обновить баланс получателя", err.Error())
	assert.Empty(t, result)
}
