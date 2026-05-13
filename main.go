package main

// D:\Project\backend_projects\audit_bot\main.go
import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/joho/godotenv"

	// Используем полное имя модуля из твоего go.mod
	"github.com/almassuleimenov/Audit_bot/bot"
	"github.com/almassuleimenov/Audit_bot/repository"
)

func main() {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Println("[WARNING] Файл .env не найден. Убедись, что он лежит рядом с main.go")
	}

	// Инициализируем токен
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("[ERROR] TELEGRAM_TOKEN not set in environment")
	}

	// Инициализируем Telegram бот
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create bot: %v", err)
	}

	log.Printf("[INFO] Authorized on account %s", botAPI.Self.UserName)

	// Инициализируем подключение к PostgreSQL
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=audit_bot port=5432 sslmode=disable TimeZone=Asia/Almaty"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to PostgreSQL: %v", err)
	}

	log.Println("[INFO] Connected to database")

	// Создаем таблицы
	err = db.AutoMigrate(&repository.AuditRecord{}, &repository.Appointment{})
	if err != nil {
		log.Fatalf("[ERROR] Failed to migrate database: %v", err)
	}

	log.Println("[INFO] Database migrations completed")

	// Инициализация слоев (Dependency Injection)
	repo := repository.NewBotRepository(db)
	handler := bot.NewBotHandler(botAPI, repo)

	// 1. Создаем функцию для обработки CORS
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Разрешаем запросы с любых доменов
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Если это предварительный запрос OPTIONS от браузера - отвечаем 200 OK
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Передаем управление дальше к самому обработчику
			next(w, r)
		}
	}

	// 2. Запускаем HTTP сервер в горутине
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		// Роут для проверки жизнеспособности сервиса (Render health check)
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Bot is alive!"))
		})

		// Эндпоинт для Аудитов (обернут в CORS)
		http.HandleFunc("/api/audits", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			var records []repository.AuditRecord
			// Запрос к БД занимает O(N)
			if err := db.Order("created_at desc").Find(&records).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(records)
		}))

		// Эндпоинт для Заявок (обернут в CORS)
		http.HandleFunc("/api/appointments", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			var appointments []repository.Appointment
			if err := db.Order("created_at desc").Find(&appointments).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(appointments)
		}))

		log.Printf("[INFO] Starting HTTP server on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("[ERROR] HTTP server failed: %v", err)
		}
	}()

	// 3. Запускаем асинхронный FSM движок (он блокирует основной поток, не давая программе завершиться)
	handler.Start()
}