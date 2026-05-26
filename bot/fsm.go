package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/almassuleimenov/Audit_bot/internal/sse"
	"github.com/almassuleimenov/Audit_bot/repository"
)

// State определяет текущий шаг пользователя в машине состояний
type State int

const (
	StateIdle State = iota
	StateWaitingBIN
	StateWaitingPosition
	StateSurveyDynamic // Единый стейт для всех динамических вопросов
	StateWaitingScore
)

// Session хранит контекст текущего диалога с пользователем
type Session struct {
	ChatID        int64
	State         State
	BIN           string
	Position      string
	Language      string // "ru" или "kk" (задел под локализацию)
	Questions     []repository.SurveyQuestion
	CurrentQIndex int
	Answers       map[string]string // Ключ: текст вопроса, Значение: ответ
	Score         int
}

// BotHandler инкапсулирует логику бота, БД и SSE
type BotHandler struct {
	bot        *tgbotapi.BotAPI
	repo       repository.BotRepository
	adminID    int64
	leadBroker *sse.Broker

	mu       sync.RWMutex
	sessions map[int64]*Session
}

// NewBotHandler конструктор для обработчика
func NewBotHandler(bot *tgbotapi.BotAPI, repo repository.BotRepository, adminID int64, broker *sse.Broker) *BotHandler {
	return &BotHandler{
		bot:        bot,
		repo:       repo,
		adminID:    adminID,
		leadBroker: broker,
		sessions:   make(map[int64]*Session),
	}
}

// getSession потокобезопасно извлекает или создает сессию
func (h *BotHandler) getSession(chatID int64) *Session {
	h.mu.RLock()
	session, exists := h.sessions[chatID]
	h.mu.RUnlock()

	if !exists {
		session = &Session{
			ChatID:   chatID,
			State:    StateIdle,
			Language: "ru", // По умолчанию ставим русский (можно расширить)
			Answers:  make(map[string]string),
		}
		h.mu.Lock()
		h.sessions[chatID] = session
		h.mu.Unlock()
	}
	return session
}

// clearSession очищает данные после завершения
func (h *BotHandler) clearSession(chatID int64) {
	h.mu.Lock()
	delete(h.sessions, chatID)
	h.mu.Unlock()
}

// Start запускает поллинг обновлений (горутина SRE-архитектуры)
func (h *BotHandler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			h.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			// Если решишь использовать Inline кнопки, обработка пойдет здесь
			h.handleCallback(update.CallbackQuery)
		}
	}
}

// handleMessage маршрутизирует входящие текстовые сообщения
func (h *BotHandler) handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	session := h.getSession(chatID)

	// Перехват глобальных команд
	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			h.clearSession(chatID)
			reply := tgbotapi.NewMessage(chatID, "Добро пожаловать в систему аудита! Отправьте команду /audit для начала проверки.")
			reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			h.bot.Send(reply)
			return
		case "audit":
			h.clearSession(chatID)
			session = h.getSession(chatID) // Создаем новую чистую сессию
			session.State = StateWaitingBIN
			
			reply := tgbotapi.NewMessage(chatID, "Пожалуйста, введите БИН вашей организации (12 цифр):")
			reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			h.bot.Send(reply)
			return
		}
	}

	// Машина состояний (FSM)
	switch session.State {
	case StateWaitingBIN:
		session.BIN = text
		session.State = StateWaitingPosition
		msg := tgbotapi.NewMessage(chatID, "Отлично. Теперь укажите вашу должность:")
		h.bot.Send(msg)

	case StateWaitingPosition:
		session.Position = text
		
		// Загружаем активные вопросы из базы данных (O(1) операция IO)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		questions, err := h.repo.GetActiveQuestions(ctx)
		if err != nil || len(questions) == 0 {
			log.Printf("[ERROR] Failed to fetch active questions: %v", err)
			msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при загрузке вопросов или вопросы отсутствуют. Попробуйте позже.")
			h.bot.Send(msg)
			h.clearSession(chatID)
			return
		}

		// Инициализируем динамический опрос
		session.Questions = questions
		session.CurrentQIndex = 0
		session.State = StateSurveyDynamic
		
		h.askCurrentQuestion(session)

	case StateSurveyDynamic:
		// Сохраняем ответ на текущий вопрос
		currentQ := session.Questions[session.CurrentQIndex]
		session.Answers[currentQ.TextRU] = text

		// Переходим к следующему
		session.CurrentQIndex++

		// Проверяем, остались ли еще вопросы
		if session.CurrentQIndex < len(session.Questions) {
			h.askCurrentQuestion(session)
		} else {
			// Вопросы закончились, переходим к финальной оценке
			session.State = StateWaitingScore
			
			msg := tgbotapi.NewMessage(chatID, "Опрос завершен! Оцените работу аудиторов по 5-балльной шкале (где 5 - отлично, 1 - очень плохо):")
			
			// Клавиатура для оценки
			row := []tgbotapi.KeyboardButton{
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("2"),
				tgbotapi.NewKeyboardButton("3"),
				tgbotapi.NewKeyboardButton("4"),
				tgbotapi.NewKeyboardButton("5"),
			}
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(row)
			h.bot.Send(msg)
		}

	case StateWaitingScore:
		score, err := strconv.Atoi(text)
		if err != nil || score < 1 || score > 5 {
			msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите число от 1 до 5 на клавиатуре.")
			h.bot.Send(msg)
			return
		}
		
		session.Score = score

		// Сохраняем все данные в БД
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		err = h.repo.SaveAuditRecord(ctx, chatID, session.BIN, session.Position, session.Answers, session.Score)
		if err != nil {
			log.Printf("[ERROR] Failed to save audit record: %v", err)
			msg := tgbotapi.NewMessage(chatID, "Произошла системная ошибка при сохранении данных. Обратитесь в поддержку.")
			h.bot.Send(msg)
			h.clearSession(chatID)
			return
		}

		msg := tgbotapi.NewMessage(chatID, "Спасибо за уделенное время! Ваши ответы успешно сохранены и помогут нам улучшить качество работы.")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		h.bot.Send(msg)

		// TODO: В следующем шаге мы будем генерировать и отправлять клиенту его уникальный номер тикета здесь.

		// Уведомляем админов через SSE (Live Dashboard)
		notification := fmt.Sprintf("Новый аудит завершен! БИН: %s, Оценка: %d", session.BIN, session.Score)
		h.leadBroker.Notifier <- []byte(notification)

		// Очищаем стейт, чтобы избежать утечек памяти
		h.clearSession(chatID)

	default:
		msg := tgbotapi.NewMessage(chatID, "Я вас не понимаю. Отправьте /audit, чтобы начать.")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		h.bot.Send(msg)
	}
}

// askCurrentQuestion формирует сообщение с текущим вопросом и динамической клавиатурой
func (h *BotHandler) askCurrentQuestion(session *Session) {
	q := session.Questions[session.CurrentQIndex]
	
	// Выбираем текст в зависимости от языка (пока хардкод на RU, можно внедрить переключатель)
	qText := q.TextRU
	var options []string
	
	// Парсим JSON варианты ответов (O(N) где N - количество вариантов ответа, обычно 2-4)
	err := json.Unmarshal(q.OptionsRU, &options)
	if err != nil {
		log.Printf("[WARNING] Не удалось распарсить варианты ответов для вопроса ID %d. Используем стандартные.", q.ID)
		options = []string{"Да", "Нет"} // Fallback механизм
	}

	msg := tgbotapi.NewMessage(session.ChatID, qText)

	// Собираем динамическую клавиатуру
	var row []tgbotapi.KeyboardButton
	for _, opt := range options {
		row = append(row, tgbotapi.NewKeyboardButton(opt))
	}
	
	// Оборачиваем строку в клавиатуру (OneTimeKeyboard для чистоты UI)
	keyboard := tgbotapi.NewReplyKeyboard(row)
	keyboard.OneTimeKeyboard = true 
	keyboard.ResizeKeyboard = true
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// handleCallback обрабатывает нажатия на Inline-кнопки (оставлено для масштабируемости)
func (h *BotHandler) handleCallback(query *tgbotapi.CallbackQuery) {
	// Подтверждаем получение коллбека, чтобы часики на кнопке перестали крутиться
	callback := tgbotapi.NewCallback(query.ID, "")
	h.bot.Request(callback)
}