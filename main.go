package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/joho/godotenv"

	"github.com/almassuleimenov/Audit_bot/bot"
	"github.com/almassuleimenov/Audit_bot/repository"
)

// parsePagination извлекает limit и offset из HTTP-запроса
func parsePagination(r *http.Request) (limit, offset int) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 50 // Дефолтное значение для защиты от выгрузки огромных данных
	}

	offset = (page - 1) * limit
	return limit, offset
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("[WARNING] Файл .env не найден. Убедись, что он лежит рядом с main.go")
	}

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("[ERROR] TELEGRAM_TOKEN not set in environment")
	}

	adminIDStr := os.Getenv("ADMIN_CHAT_ID")
	adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil || adminID == 0 {
		log.Fatal("[ERROR] ADMIN_CHAT_ID is missing or invalid in environment")
	}

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create bot: %v", err)
	}

	log.Printf("[INFO] Authorized on account %s", botAPI.Self.UserName)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=audit_bot port=5432 sslmode=disable TimeZone=Asia/Almaty"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to PostgreSQL: %v", err)
	}

	log.Println("[INFO] Connected to database")

	err = db.AutoMigrate(&repository.AuditRecord{}, &repository.Appointment{})
	if err != nil {
		log.Fatalf("[ERROR] Failed to migrate database: %v", err)
	}

	// Dependency Injection
	repo := repository.NewBotRepository(db)
	handler := bot.NewBotHandler(botAPI, repo, adminID) // Передаем Admin ID

	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next(w, r)
		}
	}

	// Настройка HTTP сервера
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		mux := http.NewServeMux()

		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Bot is alive!"))
		})

		mux.HandleFunc("/api/audits", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
			limit, offset := parsePagination(r)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			records, err := repo.GetAuditRecords(ctx, limit, offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(records)
		}))

		mux.HandleFunc("/api/appointments", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
			limit, offset := parsePagination(r)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			appointments, err := repo.GetAppointments(ctx, limit, offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(appointments)
		}))

		// Защита от Slowloris: Инициализируем сервер со строгими таймаутами
		srv := &http.Server{
			Addr:         ":" + port,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		log.Printf("[INFO] Starting HTTP server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] HTTP server failed: %v", err)
		}
	}()

	handler.Start()
}