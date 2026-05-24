package handlers

import (
	"encoding/json"
	"net/http"
		"os"
	"time"

	"github.com/almassuleimenov/Audit_bot/internal/auth" // Замени на актуальное имя модуля
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler проверяет данные и устанавливает куку
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Поиск и проверка учетных данных (в проде тут запрос к БД)
	// Сложность: O(1) поиск по индексу БД + O(1) сравнение хэшей
	if req.Username != "admin" || req.Password != "admin123" {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	tokenString, err := auth.GenerateToken(req.Username, "admin")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Отправка JWT токена в HTTP-Only куке
	// Secure флаг = true только в продакшене (по HTTPS)
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,  // Защита от XSS атак
		Secure:   true,  // Обязательно для SameSite=None, работает только по HTTPS
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}
