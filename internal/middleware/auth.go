package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/almassuleimenov/Audit_bot/internal/auth" // Замени "audit_bot" на актуальное имя твоего модуля из go.mod
)

// Определение собственного типа ключа контекста, чтобы избежать коллизий
type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware проверяет валидность токена
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value
		claims := &auth.Claims{}

		// Парсинг и криптографическая валидация токена
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return auth.JwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Прокидываем ID пользователя в контекст для последующих слоев
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}