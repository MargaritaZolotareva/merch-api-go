package handler

import (
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"merch-api/handler"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) PurchaseMerch(db *gorm.DB, username string, itemName string) (string, error) {
	args := m.Called(db, username, itemName)
	return args.String(0), args.Error(1)
}

func TestBuyItemHandler(t *testing.T) {
	mockService := new(MockService)
	mockService.On("PurchaseMerch", mock.Anything, "testuser", "item1").Return("Покупка успешна", nil)
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
	c.Params = append(c.Params, gin.Param{Key: "item", Value: "item1"})

	pHandler := handler.NewPurchaseHandler(mockService)
	pHandler.BuyItem(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Покупка успешна", response["message"])

	mockService.AssertExpectations(t)
}

func TestBuyItemHandler_UserNotAuthenticated(t *testing.T) {
	mockService := new(MockService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	pHandler := handler.NewPurchaseHandler(mockService)
	pHandler.BuyItem(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Неавторизован", response["errors"])
}

func TestBuyItemHandler_InvalidUsername(t *testing.T) {
	mockService := new(MockService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", 3.14)

	pHandler := handler.NewPurchaseHandler(mockService)
	pHandler.BuyItem(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Некорректный username", response["errors"])
}

func TestBuyItemHandler_DatabaseError(t *testing.T) {
	mockService := new(MockService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "testuser")

	pHandler := handler.NewPurchaseHandler(mockService)
	pHandler.BuyItem(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "БД недоступна", response["errors"])

	mockService.AssertExpectations(t)
}

func TestBuyItemHandler_WrongDatabaseConnection(t *testing.T) {
	mockService := new(MockService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "testuser")

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()
	c.Set("db", db)

	pHandler := handler.NewPurchaseHandler(mockService)
	pHandler.BuyItem(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Кривое подключение к БД", response["errors"])

	mockService.AssertExpectations(t)
}
