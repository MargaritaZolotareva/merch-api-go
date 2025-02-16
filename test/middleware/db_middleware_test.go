package middleware

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	middleware2 "merch-api/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDatabaseMiddleware(t *testing.T) {
	r := gin.Default()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("ошибка при открытии gorm DB: %v", err)
	}
	r.Use(middleware2.DatabaseMiddleware(gdb))

	r.GET("/test", func(c *gin.Context) {
		_, exists := c.Get("db")
		assert.True(t, exists, "db отсутствует в контексте")

		assert.IsType(t, &gorm.DB{}, c.MustGet("db"), "Тип БД должен быть *gorm.db")
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("ошибка при создании запроса: %v", err)
	}

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
