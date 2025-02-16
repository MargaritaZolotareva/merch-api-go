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

func TestGetUserInfo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать mock db: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 1000))

	mock.ExpectQuery("SELECT (.+) FROM \"purchase\" (.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"type", "quantity"}).
			AddRow("item1", 2).
			AddRow("item2", 5))

	mock.ExpectQuery("SELECT (.+) FROM \"transaction\" (.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"toUser", "amount"}).
			AddRow("user2", 100).
			AddRow("user3", 200))

	mock.ExpectQuery("SELECT (.+) FROM \"transaction\" (.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"fromUser", "amount"}).
			AddRow("user4", 150).
			AddRow("user5", 250))

	userInfoService := service2.NewUserInfoService()
	userInfo, err := userInfoService.GetUserInfo(gdb, "user1")

	assert.NoError(t, err)

	assert.Equal(t, 1000, userInfo.Coins)
	assert.Len(t, userInfo.Inventory, 2)
	assert.Equal(t, "item1", userInfo.Inventory[0].Type)
	assert.Equal(t, 2, userInfo.Inventory[0].Quantity)
	assert.Equal(t, "item2", userInfo.Inventory[1].Type)
	assert.Equal(t, 5, userInfo.Inventory[1].Quantity)
	assert.Len(t, userInfo.CoinHistory.Sent, 2)
	assert.Equal(t, "user2", userInfo.CoinHistory.Sent[0].ToUser)
	assert.Equal(t, 100, userInfo.CoinHistory.Sent[0].Amount)
	assert.Len(t, userInfo.CoinHistory.Received, 2)
	assert.Equal(t, "user4", userInfo.CoinHistory.Received[0].FromUser)
	assert.Equal(t, 150, userInfo.CoinHistory.Received[0].Amount)
}

func TestGetUserInfo_UserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать mock db: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("nonexistentUser", 1).
		WillReturnError(fmt.Errorf("пользователь не найден"))

	userInfoService := service2.NewUserInfoService()
	userInfo, err := userInfoService.GetUserInfo(gdb, "user1")

	assert.Error(t, err)
	assert.Empty(t, userInfo)
}

func TestGetUserInfo_PurchasesFetchError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать mock db: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 1000))

	mock.ExpectQuery("SELECT (.+) FROM \"purchase\" (.+)").
		WithArgs(1).
		WillReturnError(fmt.Errorf(""))

	userInfoService := service2.NewUserInfoService()
	userInfo, err := userInfoService.GetUserInfo(gdb, "user1")

	assert.Error(t, err)
	assert.Equal(t, "не удалось получить инвентарь пользователя: ", err.Error())
	assert.Empty(t, userInfo.Inventory)
}

func TestGetUserInfo_SentTransactionsFetchError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать mock db: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 1000))

	mock.ExpectQuery("SELECT (.+) FROM \"purchase\" (.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"type", "quantity"}).
			AddRow("item1", 2).
			AddRow("item2", 5))

	mock.ExpectQuery("SELECT (.+) FROM \"transaction\" (.+)").
		WithArgs(1).
		WillReturnError(fmt.Errorf(""))

	userInfoService := service2.NewUserInfoService()
	userInfo, err := userInfoService.GetUserInfo(gdb, "user1")

	assert.Error(t, err)
	assert.Equal(t, "не удалось получить отправленные транзакции пользователя: ", err.Error())
	assert.Empty(t, userInfo.CoinHistory.Sent)
}

func TestGetUserInfo_ReceivedTransactionsFetchError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать mock db: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}

	mock.ExpectQuery("SELECT (.+) FROM \"employee\" WHERE username = (.+)").
		WithArgs("user1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "balance"}).
			AddRow(1, "user1", 1000))

	mock.ExpectQuery("SELECT (.+) FROM \"purchase\" (.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"type", "quantity"}).
			AddRow("item1", 2).
			AddRow("item2", 5))

	mock.ExpectQuery("SELECT (.+) FROM \"transaction\" (.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"toUser", "amount"}).
			AddRow("user2", 100).
			AddRow("user3", 200))

	mock.ExpectQuery("SELECT (.+) FROM \"transaction\" (.+)").
		WithArgs(1).
		WillReturnError(fmt.Errorf(""))

	userInfoService := service2.NewUserInfoService()
	userInfo, err := userInfoService.GetUserInfo(gdb, "user1")

	assert.Error(t, err)
	assert.Equal(t, "не удалось получить полученные транзакции пользователя: ", err.Error())
	assert.Empty(t, userInfo.CoinHistory.Received)
}
