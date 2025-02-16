package e2e

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	model2 "merch-api/model"
	router2 "merch-api/router"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendCoins_E2E(t *testing.T) {
	resetTables()
	router := router2.SetupRouter(db)
	hashedPswd1, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	hashedPswd2, err := bcrypt.GenerateFromPassword([]byte("password456"), bcrypt.MinCost)

	db.Create(&model2.Employee{
		Username: "test_user1",
		Password: string(hashedPswd1),
		Balance:  100,
	})
	db.Create(&model2.Employee{
		Username: "test_user2",
		Password: string(hashedPswd2),
		Balance:  100,
	})
	merchName := "cup"

	authRequestBody, _ := json.Marshal(map[string]string{"username": "test_user1", "password": "password123"})
	authReq := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewBuffer(authRequestBody))
	authReq.Header.Set("Content-Type", "application/json")

	authW := httptest.NewRecorder()
	router.ServeHTTP(authW, authReq)
	assert.Equal(t, http.StatusOK, authW.Code)
	var authResponse map[string]interface{}
	err = json.NewDecoder(authW.Body).Decode(&authResponse)
	assert.NoError(t, err)
	token := authResponse["token"].(string)

	requestBody, _ := json.Marshal(map[string]interface{}{"toUser": "test_user2", "amount": 30})
	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Перевод успешен")

	var employee1 model2.Employee
	db.First(&employee1, "username = ?", "test_user1")
	assert.Equal(t, 70, employee1.Balance)

	var employee2 model2.Employee
	db.First(&employee2, "username = ?", "test_user2")
	assert.Equal(t, 130, employee2.Balance)

	var transaction model2.Transaction
	var merch model2.Merch
	db.First(&merch, "name = ?", merchName)
	db.First(&transaction, "sender_id = ? AND receiver_id = ?", employee1.ID, employee2.ID)
	assert.NotNil(t, transaction)
}
