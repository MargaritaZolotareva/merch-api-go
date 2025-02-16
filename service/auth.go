package service

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"merch-api/model"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("JWT_KEY"))
var (
	ErrInvalidInput          = fmt.Errorf("invalid input")
	ErrPasswordMismatch      = fmt.Errorf("invalid password")
	ErrFailedToCreateUser    = fmt.Errorf("failed to find or create employee")
	ErrFailedToGenerateToken = fmt.Errorf("failed to generate token")
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type AuthService interface {
	AuthenticateUser(db *gorm.DB, username, password string) (string, error)
}

type AuthServiceImpl struct{}

func NewAuthService() *AuthServiceImpl {
	return &AuthServiceImpl{}
}

func GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("ошибка при подписании токена: %v", err)
	}

	return tokenString, nil
}

func (s *AuthServiceImpl) AuthenticateUser(db *gorm.DB, username, password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	employee, err := findOrCreateEmployee(db, username, string(hashedPassword))
	if err != nil {
		return "", ErrFailedToCreateUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(employee.Password), []byte(password)); err != nil {
		return "", ErrPasswordMismatch
	}

	token, err := GenerateJWT(employee.Username)
	if err != nil {
		return "", ErrFailedToGenerateToken
	}

	return token, nil
}

func findOrCreateEmployee(db *gorm.DB, username, hashedPassword string) (*model.Employee, error) {
	var employee model.Employee
	if err := db.Where("username = ?", username).
		Attrs(model.Employee{Username: username, Password: hashedPassword, Balance: 1000}).
		FirstOrCreate(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}
