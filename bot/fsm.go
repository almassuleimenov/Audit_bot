package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/almassuleimenov/Audit_bot/internal/sse"
	"github.com/almassuleimenov/Audit_bot/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	mu            sync.Mutex // Блокировка состояния конкретной сессии
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

// Reset зануляет состояние FSM за O(1).
// Избавляет от необходимости удалять структуру из глобальной map и дергать RW мьютексы.
func (s *Session) Reset() {
	s.State = StateIdle
	s.BIN = ""
	s.Position = ""
	s.Questions = nil
	s.CurrentQIndex = 0
	s.Answers = make(map[string]string)
	s.Score = 0
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

// sendMainMenu возвращает пользователя в начальную точку FSM
func (h *BotHandler) sendMainMenu(chatID int64) {
	text := "Вас приветствует Чат-бот Департамента внутреннего государственного аудита.\n\nВыберите действие:"
	msg := tgbotapi.NewMessage(chatID, text)

	btn := tgbotapi.NewKeyboardButton("/audit")
	row := []tgbotapi.KeyboardButton{btn}
	keyboard := tgbotapi.NewReplyKeyboard(row)
	keyboard.ResizeKeyboard = true
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

func (h *BotHandler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for update := range updates {
		// Асинхронная O(1) диспетчеризация
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

	// Захватываем контроль над сессией на время обработки команды
	session.mu.Lock()
	defer session.mu.Unlock()

	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			session.Reset()
			h.sendMainMenu(chatID)
			return
		case "audit":
			session.Reset()
			session.State = StateWaitingBIN

			reply := tgbotapi.NewMessage(chatID, "Пожалуйста, введите БИН вашей организации (12 цифр):")
			reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			h.bot.Send(reply)
			return
		}
	}

	switch session.State {
	case StateWaitingBIN:
		// Базовая пре-валидация
		if len(text) != 12 {
			reply := tgbotapi.NewMessage(chatID, "БИН должен состоять из 12 символов. Попробуйте снова:")
			h.bot.Send(reply)
			return
		}

		session.BIN = text
		session.State = StateWaitingPosition
		reply := tgbotapi.NewMessage(chatID, "Отлично. Теперь укажите вашу должность:")
		h.bot.Send(reply)

	case StateWaitingPosition:
		session.Position = text

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		questions, err := h.repo.GetActiveQuestions(ctx)
		if err != nil || len(questions) == 0 {
			log.Printf("[ERROR] Failed to fetch active questions: %v", err)
			reply := tgbotapi.NewMessage(chatID, "Произошла ошибка при загрузке вопросов или они отсутствуют. Попробуйте позже.")
			h.bot.Send(reply)
			
			// Возврат в меню при системной ошибке
			session.Reset()
			h.sendMainMenu(chatID)
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

			reply := tgbotapi.NewMessage(chatID, "Анкета завершена. Оцените работу аудиторов по 5-балльной шкале (где 5 - отлично, 1 - очень плохо):")
			row := []tgbotapi.KeyboardButton{
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("2"),
				tgbotapi.NewKeyboardButton("3"),
				tgbotapi.NewKeyboardButton("4"),
				tgbotapi.NewKeyboardButton("5"),
			}
			keyboard := tgbotapi.NewReplyKeyboard(row)
			keyboard.ResizeKeyboard = true
			reply.ReplyMarkup = keyboard
			h.bot.Send(reply)
		}

	case StateWaitingScore:
		score, err := strconv.Atoi(text)
		if err != nil || score < 1 || score > 5 {
			reply := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите число от 1 до 5 на клавиатуре.")
			h.bot.Send(reply)
			return
		}

		session.Score = score

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = h.repo.SaveAuditRecord(ctx, chatID, session.BIN, session.Position, session.Answers, session.Score)
		if err != nil {
			log.Printf("[ERROR] Failed to save audit record: %v", err)
			reply := tgbotapi.NewMessage(chatID, "Произошла системная ошибка при сохранении данных. Обратитесь в поддержку.")
			h.bot.Send(reply)
			session.Reset()
			h.sendMainMenu(chatID)
			return
		}

		reply := tgbotapi.NewMessage(chatID, "Спасибо за уделенное время! Ваши ответы успешно сохранены и помогут нам улучшить качество работы.")
		reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		h.bot.Send(reply)

		notification := fmt.Sprintf("Новый аудит завершен! БИН: %s, Оценка: %d", session.BIN, session.Score)
		// Неблокирующая отправка в канал
		select {
		case h.leadBroker.Notifier <- []byte(notification):
		default:
			log.Println("[WARNING] SSE Notifier channel is full, skipping broadcast")
		}

		// Замыкание FSM: Сбрасываем контекст и выдаем главное меню
		session.Reset()
		h.sendMainMenu(chatID)

	default:
		// Fallback если стейт пуст, а юзер прислал рандомный текст
		session.Reset()
		h.sendMainMenu(chatID)
	}
}

func (h *BotHandler) askCurrentQuestion(session *Session) {
	q := session.Questions[session.CurrentQIndex]
	qText := fmt.Sprintf("%d. %s", session.CurrentQIndex+1, q.TextRU)
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