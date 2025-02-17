package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"merch-api/middleware"
	"merch-api/service"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestJWTMiddleware_NoToken(t *testing.T) {
	r := gin.Default()
	r.Use(middleware.JWTMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("ошибка при создании запроса: %v", err)
	}

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("ошибка при декодировании ответа: %v", err)
	}
	assert.Equal(t, "Authorization token is required", response["errors"])
}

func TestJWTMiddleware_InvalidTokenFormat(t *testing.T) {
	r := gin.Default()
	r.Use(middleware.JWTMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("ошибка при создании запроса: %v", err)
	}
	req.Header.Set("Authorization", "InvalidTokenString")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("ошибка при декодировании ответа: %v", err)
	}
	assert.Equal(t, "Кривой формат токена", response["errors"])
}

func TestJWTMiddleware_InvalidOrExpiredToken(t *testing.T) {
	invalidToken := "InvalidTokenString"

	r := gin.Default()
	r.Use(middleware.JWTMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("ошибка при создании запроса: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+invalidToken)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("ошибка при декодировании ответа: %v", err)
	}
	assert.Equal(t, "Invalid or expired token", response["errors"])
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	tokenString, err := generateValidToken("testuser")
	if err != nil {
		t.Fatalf("ошибка при генерации валидного токена: %v", err)
	}

	r := gin.Default()
	r.Use(middleware.JWTMiddleware())
	r.GET("/test", func(c *gin.Context) {
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{"message": "Success", "username": username})
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("ошибка при создании запроса: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("ошибка при декодировании ответа: %v", err)
	}
	fmt.Println(response)
	assert.Equal(t, "Success", response["message"])
	assert.Equal(t, "testuser", response["username"])
}

func generateValidToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := service.Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("ошибка при подписании токена: %v", err)
	}

	return tokenString, nil
}
