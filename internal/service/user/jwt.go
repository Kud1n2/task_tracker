package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret string
}

func NewJWT(secret string) *JWTManager {
	return &JWTManager{
		secret: secret,
	}
}

func (j *JWTManager) Generate(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(2 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(j.secret))
}
