package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	model2 "merch-api/model"
	router2 "merch-api/router"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Ошибка при получении рабочего каталога: %v", err)
	}

	envFile := filepath.Join(wd, "../..", ".env.test")
	err = godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Ошибка при загрузке .env файла: %v", err)
	}
	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to db: " + err.Error())
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	code := m.Run()
	os.Exit(code)
}

func resetTables() {
	db.Exec("TRUNCATE TABLE employee, transaction, purchase RESTART IDENTITY CASCADE;")
}

func TestPurchaseMerch_E2E(t *testing.T) {
	resetTables()
	router := router2.SetupRouter(db)
	hashedPswd, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	merchName := "cup"

	db.Create(&model2.Employee{
		Username: "test_user1",
		Password: string(hashedPswd),
		Balance:  100,
	})
	db.FirstOrCreate(&model2.Merch{
		Name:  merchName,
		Price: 20,
	})

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

	req := httptest.NewRequest(http.MethodGet, "/api/buy/"+merchName, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Покупка успешна")

	var employee model2.Employee
	db.First(&employee, "username = ?", "test_user1")
	assert.Equal(t, 80, employee.Balance)

	var purchase model2.Purchase
	var merch model2.Merch
	db.First(&merch, "name = ?", merchName)
	db.First(&purchase, "employee_id = ? AND merch_id = ?", employee.ID, merch.ID)
	assert.NotNil(t, purchase)
}
