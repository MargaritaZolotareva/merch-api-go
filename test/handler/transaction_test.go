package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) SendCoins(db *gorm.DB, fromUsername, toUsername string, amount int) (string, error) {
	args := m.Called(db, fromUsername, toUsername, amount)
	return args.String(0), args.Error(1)
}

func TestSendCoinHandler(t *testing.T) {
	mockService := new(MockTransactionService)
	mockService.On("SendCoins", mock.Anything, "testuser1", "testuser2", 100).Return(fmt.Sprintf("Перевод успешен! Кол-во: %d монет пользователю %s.", 100, "testuser2"), nil)
	requestBody := map[string]interface{}{
		"toUser": "testuser2",
		"amount": 100,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Ошибка при маршализации тела запроса: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "testuser1")
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
	c.Request, _ = http.NewRequest(http.MethodPost, "/sendCoin", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	tHandler := handler.NewTransactionHandler(mockService)
	tHandler.SendCoin(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Перевод успешен! Кол-во: 100 монет пользователю testuser2.", response["message"])

	mockService.AssertExpectations(t)
}

func TestSendCoin_UserNotAuthenticated(t *testing.T) {
	mockService := new(MockTransactionService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	tHandler := handler.NewTransactionHandler(mockService)
	tHandler.SendCoin(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Неавторизован", response["errors"])
}

func TestSendCoin_InvalidUsername(t *testing.T) {
	mockService := new(MockTransactionService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", 12345)

	tHandler := handler.NewTransactionHandler(mockService)
	tHandler.SendCoin(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Некорректный username", response["errors"])
}

func TestSendCoinHandler_InvalidRequest(t *testing.T) {
	mockService := new(MockTransactionService)
	requestBody := map[string]interface{}{
		"toUser": "testuser2",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Ошибка при маршализации тела запроса: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

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
	c.Set("username", "testuser1")

	c.Request, _ = http.NewRequest(http.MethodPost, "/sendCoin", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	uHandler := handler.NewTransactionHandler(mockService)
	uHandler.SendCoin(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Неверный запрос", response["errors"])

	mockService.AssertExpectations(t)
}

func TestSendCoin_SelfTransfer(t *testing.T) {
	mockService := new(MockTransactionService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "testuser1")

	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/sendCoin",
		bytes.NewBufferString(`{"toUser": "testuser1", "amount": 100}`),
	)

	tHandler := handler.NewTransactionHandler(mockService)
	tHandler.SendCoin(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Нельзя отправить монеты самому себе", response["errors"])
}

func TestSendCoin_DatabaseNotFound(t *testing.T) {
	mockService := new(MockTransactionService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "testuser")
	c.Request = httptest.NewRequest(http.MethodPost, "/sendCoin", bytes.NewBufferString(`{"toUser": "testuser2", "amount": 100}`))

	tHandler := handler.NewTransactionHandler(mockService)
	tHandler.SendCoin(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "БД недоступна", response["errors"])
}

func TestSendCoin_WrongDatabaseConnection(t *testing.T) {
	mockService := new(MockTransactionService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "testuser1")

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()
	c.Set("db", db)
	c.Request = httptest.NewRequest(http.MethodPost, "/sendCoin", bytes.NewBufferString(`{"toUser": "testuser2", "amount": 100}`))

	tHandler := handler.NewTransactionHandler(mockService)
	tHandler.SendCoin(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Кривое подключение к БД", response["errors"])
}
