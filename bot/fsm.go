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

type State int

const (
	StateIdle State = iota
	StateWaitingBIN
	StateWaitingPosition
	StateSurveyDynamic
	StateWaitingScore
)

type Session struct {
	mu            sync.Mutex // Блокировка состояния конкретной сессии для защиты map и полей
	ChatID        int64
	State         State
	BIN           string
	Position      string
	Language      string
	Questions     []repository.SurveyQuestion
	CurrentQIndex int
	Answers       map[string]string
	Score         int
}

type BotHandler struct {
	bot        *tgbotapi.BotAPI
	repo       repository.BotRepository
	adminID    int64
	leadBroker *sse.Broker

	mu       sync.RWMutex // Блокировка только для пула сессий
	sessions map[int64]*Session
}

func NewBotHandler(bot *tgbotapi.BotAPI, repo repository.BotRepository, adminID int64, broker *sse.Broker) *BotHandler {
	return &BotHandler{
		bot:        bot,
		repo:       repo,
		adminID:    adminID,
		leadBroker: broker,
		sessions:   make(map[int64]*Session),
	}
}

// getSession использует паттерн Double-Checked Locking
func (h *BotHandler) getSession(chatID int64) *Session {
	h.mu.RLock()
	session, exists := h.sessions[chatID]
	h.mu.RUnlock()

	if !exists {
		h.mu.Lock()
		defer h.mu.Unlock()
		// Двойная проверка на случай, если другая горутина уже создала сессию
		session, exists = h.sessions[chatID]
		if !exists {
			session = &Session{
				ChatID:   chatID,
				State:    StateIdle,
				Language: "ru",
				Answers:  make(map[string]string),
			}
			h.sessions[chatID] = session
		}
	}
	return session
}

func (h *BotHandler) clearSession(chatID int64) {
	h.mu.Lock()
	delete(h.sessions, chatID)
	h.mu.Unlock()
}

func (h *BotHandler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for update := range updates {
		// Асинхронная обработка O(1) диспетчеризация
		if update.Message != nil {
			go h.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			go h.handleCallback(update.CallbackQuery)
		}
	}
}

func (h *BotHandler) handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	session := h.getSession(chatID)
	
	// Блокируем конкретную сессию на время обработки сообщения.
	// Защищает от спама и Data Races внутри State Machine.
	session.mu.Lock()
	defer session.mu.Unlock()

	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			// Так как мы внутри лока сессии, вызываем очистку пула безопасно 
			// (в clearSession свой лок на мапу sessions)
			h.clearSession(chatID)
			reply := tgbotapi.NewMessage(chatID, "Добро пожаловать в систему аудита! Отправьте команду /audit для начала проверки.")
			reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			h.bot.Send(reply)
			return
		case "audit":
			h.clearSession(chatID)
			
			// Мы очистили сессию в мапе, но текущая горутина все еще держит мьютекс старого объекта session.
			// Необходимо отпустить его и получить новый.
			session.mu.Unlock()
			session = h.getSession(chatID)
			session.mu.Lock()
			
			session.State = StateWaitingBIN
			
			reply := tgbotapi.NewMessage(chatID, "Пожалуйста, введите БИН вашей организации (12 цифр):")
			reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			h.bot.Send(reply)
			return
		}
	}

	switch session.State {
	case StateWaitingBIN:
		session.BIN = text
		session.State = StateWaitingPosition
		msg := tgbotapi.NewMessage(chatID, "Отлично. Теперь укажите вашу должность:")
		h.bot.Send(msg)

	case StateWaitingPosition:
		session.Position = text
		
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

		session.Questions = questions
		session.CurrentQIndex = 0
		session.State = StateSurveyDynamic
		
		h.askCurrentQuestion(session)

	case StateSurveyDynamic:
		currentQ := session.Questions[session.CurrentQIndex]
		session.Answers[currentQ.TextRU] = text
		session.CurrentQIndex++

		if session.CurrentQIndex < len(session.Questions) {
			h.askCurrentQuestion(session)
		} else {
			session.State = StateWaitingScore
			
			msg := tgbotapi.NewMessage(chatID, "Опрос завершен! Оцените работу аудиторов по 5-балльной шкале (где 5 - отлично, 1 - очень плохо):")
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

		notification := fmt.Sprintf("Новый аудит завершен! БИН: %s, Оценка: %d", session.BIN, session.Score)
		h.leadBroker.Notifier <- []byte(notification)

		h.clearSession(chatID)

	default:
		msg := tgbotapi.NewMessage(chatID, "Я вас не понимаю. Отправьте /audit, чтобы начать.")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		h.bot.Send(msg)
	}
}

func (h *BotHandler) askCurrentQuestion(session *Session) {
	q := session.Questions[session.CurrentQIndex]
	qText := q.TextRU
	var options []string
	
	err := json.Unmarshal(q.OptionsRU, &options)
	if err != nil {
		log.Printf("[WARNING] Не удалось распарсить варианты ответов для вопроса ID %d. Используем стандартные.", q.ID)
		options = []string{"Да", "Нет"}
	}

	msg := tgbotapi.NewMessage(session.ChatID, qText)
	var row []tgbotapi.KeyboardButton
	for _, opt := range options {
		row = append(row, tgbotapi.NewKeyboardButton(opt))
	}
	
	keyboard := tgbotapi.NewReplyKeyboard(row)
	keyboard.OneTimeKeyboard = true 
	keyboard.ResizeKeyboard = true
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

func (h *BotHandler) handleCallback(query *tgbotapi.CallbackQuery) {
	callback := tgbotapi.NewCallback(query.ID, "")
	h.bot.Request(callback)
}