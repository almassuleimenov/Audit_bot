package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	// 1. Подключаем новую библиотеку для работы с файлом .env
	"github.com/joho/godotenv"

	// Используем полное имя модуля из твоего go.mod
	"github.com/yourusername/audit_bot/bot"
	"github.com/yourusername/audit_bot/repository"
)

func main() {
	// 2. Вызываем функцию Load(). Она автоматически найдет файл .env в папке
	// и загрузит из него TELEGRAM_TOKEN в системные переменные программы.
	err := godotenv.Load()
	if err != nil {
		// Если файла нет, программа не упадет, а просто выведет предупреждение.
		log.Println("[WARNING] Файл .env не найден. Убедись, что он лежит рядом с main.go")
	}

	// 3. Загружаем токен. Теперь программа найдет его без проблем!
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
	// Настрой user, password и dbname под свою локальную БД
	dsn := "host=localhost user=postgres password=postgres dbname=audit_bot port=5432 sslmode=disable TimeZone=Asia/Almaty"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to PostgreSQL: %v", err)
	}

	log.Println("[INFO] Connected to database")

	// Создаем таблицы, если их нет
	err = db.AutoMigrate(&repository.AuditRecord{}, &repository.Appointment{})
	if err != nil {
		log.Fatalf("[ERROR] Failed to migrate database: %v", err)
	}

	log.Println("[INFO] Database migrations completed")

	// Инициализация слоев (Dependency Injection)
	repo := repository.NewBotRepository(db)
	handler := bot.NewBotHandler(botAPI, repo)

	// Запускаем асинхронный FSM движок
	handler.Start()
}