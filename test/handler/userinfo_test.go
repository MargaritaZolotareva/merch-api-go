package handler

import (
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	handler2 "merch-api/handler"
	"merch-api/service"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockUserInfoService struct {
	mock.Mock
}

func (m *MockUserInfoService) GetUserInfo(db *gorm.DB, username string) (service.UserInfo, error) {
	args := m.Called(db, username)
	return args.Get(0).(service.UserInfo), args.Error(1)
}

func TestInfoHandler(t *testing.T) {
	mockService := new(MockUserInfoService)
	mockService.On("GetUserInfo", mock.Anything, "testuser").Return(service.UserInfo{
		Coins: 1000,
		Inventory: []service.InventoryItem{
			{"item1", 5},
			{"item2", 10},
		},
		CoinHistory: service.CoinHistoryItem{
			Received: []service.ReceivedCoinsItem{
				{"user1", 20},
			},
			Sent: []service.SentCoinsItem{
				{"user2", 30},
			},
		},
	}, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "testuser")

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}
	c.Set("db", gdb)

	uHandler := handler2.NewUserInfoHandler(mockService)
	uHandler.InfoHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response service.UserInfo
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 1000, response.Coins)
	assert.Len(t, response.Inventory, 2)
	assert.Equal(t, "item1", response.Inventory[0].Type)
	assert.Equal(t, 5, response.Inventory[0].Quantity)
	assert.Equal(t, "item2", response.Inventory[1].Type)
	assert.Equal(t, 10, response.Inventory[1].Quantity)
	assert.Len(t, response.CoinHistory.Received, 1)
	assert.Equal(t, "user1", response.CoinHistory.Received[0].FromUser)
	assert.Equal(t, 20, response.CoinHistory.Received[0].Amount)
	assert.Len(t, response.CoinHistory.Sent, 1)
	assert.Equal(t, "user2", response.CoinHistory.Sent[0].ToUser)
	assert.Equal(t, 30, response.CoinHistory.Sent[0].Amount)

	mockService.AssertExpectations(t)
}

func TestInfoHandler_UserNotAuthenticated(t *testing.T) {
	mockService := new(MockUserInfoService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	uHandler := handler2.NewUserInfoHandler(mockService)
	uHandler.InfoHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Неавторизован", response["errors"])
}

func TestInfoHandler_InvalidUsername(t *testing.T) {
	mockService := new(MockUserInfoService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", 12345)

	uHandler := handler2.NewUserInfoHandler(mockService)
	uHandler.InfoHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Некорректный username", response["errors"])
}

func TestInfoHandler_DatabaseNotFound(t *testing.T) {
	mockService := new(MockUserInfoService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "testuser")

	uHandler := handler2.NewUserInfoHandler(mockService)
	uHandler.InfoHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "БД недоступна", response["errors"])
}

func TestInfoHandler_WrongDatabaseConnection(t *testing.T) {
	mockService := new(MockUserInfoService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "testuser")

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()
	c.Set("db", db)

	uHandler := handler2.NewUserInfoHandler(mockService)
	uHandler.InfoHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Кривое подключение к БД", response["errors"])
}
