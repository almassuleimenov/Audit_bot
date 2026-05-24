package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// В production этот ключ должен читаться из os.Getenv("JWT_SECRET")
var JwtSecretKey = []byte("SUPER_SECRET_KEY_REPLACE_IN_PROD")

// Claims описывает структуру полезной нагрузки
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken создает подписанный токен
func GenerateToken(userID, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecretKey)
}