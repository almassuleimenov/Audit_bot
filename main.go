package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/almassuleimenov/Audit_bot/bot"
	"github.com/almassuleimenov/Audit_bot/internal/handlers"
	"github.com/almassuleimenov/Audit_bot/internal/middleware"
	"github.com/almassuleimenov/Audit_bot/internal/sse"
	"github.com/almassuleimenov/Audit_bot/repository"
)

func parsePagination(r *http.Request) (limit, offset int) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	offset = (page - 1) * limit
	return limit, offset
}

// corsMiddleware для обычных HandlerFunc (используется для логина)
func corsMiddleware(allowedOrigin string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// corsMiddlewareForHandler для интерфейса http.Handler (используется в цепочке с JWT)
func corsMiddlewareForHandler(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	_ = godotenv.Load()

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("[ERROR] TELEGRAM_TOKEN not set in environment")
	}

	adminIDStr := os.Getenv("ADMIN_CHAT_ID")
	adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil || adminID == 0 {
		log.Fatal("[ERROR] ADMIN_CHAT_ID is missing or invalid")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("[ERROR] DATABASE_URL is required in environment variables")
	}

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		log.Fatal("[ERROR] ALLOWED_ORIGIN is required (e.g. http://localhost:3000)")
	}

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create bot: %v", err)
	}

	log.Printf("[INFO] Authorized on account %s", botAPI.Self.UserName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to PostgreSQL: %v", err)
	}

	err = db.AutoMigrate(&repository.AuditRecord{}, &repository.Appointment{}, &repository.SurveyQuestion{})
	if err != nil {
		log.Fatalf("[ERROR] Failed to migrate database: %v", err)
	}

	repo := repository.NewBotRepository(db)

	if err := repo.SeedDefaultQuestions(context.Background()); err != nil {
		log.Printf("[WARNING] Ошибка заполнения вопросов: %v", err)
	}

	// Инициализируем брокер SSE
	leadBroker := sse.NewBroker()
	go leadBroker.Start()

	handler := bot.NewBotHandler(botAPI, repo, adminID, leadBroker)

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		mux := http.NewServeMux()

		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Bot is alive!"))
		})

		// НОВАЯ ОБЕРТКА ДЛЯ АВТОРИЗАЦИИ
		// Объединяет CORS и JWT Middleware. Принимает стандартный HandlerFunc.
		protectJWT := func(h http.HandlerFunc) http.Handler {
			return corsMiddlewareForHandler(allowedOrigin, middleware.AuthMiddleware(http.HandlerFunc(h)))
		}

		// Обрати внимание: теперь мы используем mux.Handle вместо mux.HandleFunc для защищенных роутов
		mux.Handle("/api/audits", protectJWT(func(w http.ResponseWriter, r *http.Request) {
			limit, offset := parsePagination(r)
			records, err := repo.GetAuditRecords(context.Background(), limit, offset)
			if err != nil {
				http.Error(w, "Failed to get audit records", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(records)
		}))

		mux.Handle("/api/appointments", protectJWT(func(w http.ResponseWriter, r *http.Request) {
			limit, offset := parsePagination(r)
			appointments, err := repo.GetAppointments(context.Background(), limit, offset)
			if err != nil {
				http.Error(w, "Failed to get appointments", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(appointments)
		}))

		mux.Handle("/api/questions", protectJWT(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == "GET" {
				questions, err := repo.GetAllQuestions(context.Background())
				if err != nil {
					http.Error(w, "Failed to retrieve questions", http.StatusInternalServerError)
					return
				}
				_ = json.NewEncoder(w).Encode(questions)
				return
			}
			if r.Method == "POST" {
				var q repository.SurveyQuestion
				if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
					http.Error(w, "Invalid payload", http.StatusBadRequest)
					return
				}
				if err := repo.SaveQuestion(context.Background(), &q); err != nil {
					http.Error(w, "Failed to save question", http.StatusInternalServerError)
					return
				}
				_ = json.NewEncoder(w).Encode(q)
				return
			}
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}))

		mux.Handle("/api/export", protectJWT(func(w http.ResponseWriter, r *http.Request) {
			var records []repository.AuditRecord
			if err := db.Order("created_at desc").Find(&records).Error; err != nil {
				http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
				return
			}

			f := excelize.NewFile()
			defer func() {
				if err := f.Close(); err != nil {
					log.Printf("[ERROR] Failed to close excelize file: %v", err)
				}
			}()

			sheetData := "Данные"
			sheetSummary := "Сводка (Анализ)"

			_ = f.SetSheetName("Sheet1", sheetData)
			if _, err := f.NewSheet(sheetSummary); err != nil {
				http.Error(w, "Failed to create summary sheet", http.StatusInternalServerError)
				return
			}

			headers := []interface{}{"ID", "Дата", "БИН", "Должность", "Оценка", "Ответы (JSON)"}
			_ = f.SetSheetRow(sheetData, "A1", &headers)

			var sumScore int
			var count5, count4, countBad int

			for i, rec := range records {
				rowNum := i + 2
				sumScore += rec.Score
				if rec.Score == 5 {
					count5++
				} else if rec.Score == 4 {
					count4++
				} else {
					countBad++
				}

				row := []interface{}{
					rec.ID,
					rec.CreatedAt.Format("02.01.2006 15:04"),
					rec.BIN,
					rec.Position,
					rec.Score,
					string(rec.Answers),
				}
				_ = f.SetSheetRow(sheetData, fmt.Sprintf("A%d", rowNum), &row)
			}

			_ = f.SetColWidth(sheetData, "A", "A", 5)
			_ = f.SetColWidth(sheetData, "B", "D", 20)
			_ = f.SetColWidth(sheetData, "E", "E", 10)
			_ = f.SetColWidth(sheetData, "F", "F", 50)

			total := len(records)
			avgScore := 0.0
			if total > 0 {
				avgScore = float64(sumScore) / float64(total)
			}

			summaryHeaders := []interface{}{"Метрика", "Значение"}
			_ = f.SetSheetRow(sheetSummary, "A1", &summaryHeaders)

			summaryData := [][]interface{}{
				{"Всего проведенных аудитов", total},
				{"Средний балл", fmt.Sprintf("%.2f", avgScore)},
				{"Количество оценок '5'", count5},
				{"Количество оценок '4'", count4},
				{"Оценки '3' и ниже (Риск)", countBad},
			}

			for i, row := range summaryData {
				_ = f.SetSheetRow(sheetSummary, fmt.Sprintf("A%d", i+2), &row)
			}

			_ = f.SetColWidth(sheetSummary, "A", "A", 30)
			_ = f.SetColWidth(sheetSummary, "B", "B", 15)

			idx, err := f.GetSheetIndex(sheetSummary)
			if err == nil {
				f.SetActiveSheet(idx)
			}

			w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=Audit_Report_%s.xlsx", time.Now().Format("2006-01-02")))

			if err := f.Write(w); err != nil {
				log.Printf("[ERROR] Excel export write failed: %v", err)
			}
		}))

		mux.Handle("/api/stats", protectJWT(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			type StatsResponse struct {
				TotalAudits       int64            `json:"total_audits"`
				AverageScore      float64          `json:"average_score"`
				TotalAppointments int64            `json:"total_appointments"`
				ScoreDistribution map[string]int64 `json:"score_distribution"`
				DailyDynamics     []struct {
					Date  string `json:"name"`
					Count int64  `json:"Проверки"`
				} `json:"daily_dynamics"`
			}

			var stats StatsResponse
			stats.ScoreDistribution = make(map[string]int64)
			if err := db.Model(&repository.AuditRecord{}).
				Select("TO_CHAR(created_at, 'DD Mon') as date, count(*) as count").
				Group("TO_CHAR(created_at, 'DD Mon')").
				Order("MIN(created_at) ASC").
				Scan(&stats.DailyDynamics).Error; err != nil {
				log.Printf("[ERROR] Failed to scan daily dynamics: %v", err)
			}

			if err := db.Model(&repository.AuditRecord{}).Count(&stats.TotalAudits).Error; err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			if err := db.Model(&repository.Appointment{}).Count(&stats.TotalAppointments).Error; err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}

			if stats.TotalAudits > 0 {
				if err := db.Model(&repository.AuditRecord{}).Select("COALESCE(AVG(score), 0)").Scan(&stats.AverageScore).Error; err != nil {
					log.Printf("[ERROR] Failed to scan average score: %v", err)
				}

				var distribution []struct {
					Score int
					Count int64
				}
				if err := db.Model(&repository.AuditRecord{}).Select("score, count(*) as count").Group("score").Scan(&distribution).Error; err != nil {
					log.Printf("[ERROR] Failed to scan score distribution: %v", err)
				}

				for _, d := range distribution {
					stats.ScoreDistribution[strconv.Itoa(d.Score)] = d.Count
				}
			}

			_ = json.NewEncoder(w).Encode(stats)
		}))

		// Публичный роут логина
		mux.HandleFunc("/api/login", corsMiddleware(allowedOrigin, handlers.LoginHandler))

		// SSE роут с единым JWT middleware
		mux.Handle("/api/stream/leads", corsMiddlewareForHandler(allowedOrigin, middleware.AuthMiddleware(leadBroker)))

		srv := &http.Server{
			Addr:         ":" + port,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		log.Printf("[INFO] Starting HTTP server on port %s", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] HTTP server failed: %v", err)
		}
	}()

	handler.Start()
}