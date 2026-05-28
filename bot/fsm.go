package bot
//D:\Project\backend_projects\audit_bot\bot\fsm.go
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
	StateWaitingLanguage
	StateWaitingPhone
	StateWaitingBIN
	StateWaitingPosition
	StateSurveyDynamic
	StateWaitingScore
)

type Session struct {
	mu            sync.Mutex
	ChatID        int64
	State         State
	PhoneNumber   string
	BIN           string
	Position      string
	Language      string
	Questions     []repository.SurveyQuestion
	CurrentQIndex int
	Answers       map[string]string
	Score         int
}

func (s *Session) Reset() {
	s.State = StateIdle
	s.PhoneNumber = ""
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

	mu       sync.RWMutex
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
				Language: "",
				Answers:  make(map[string]string),
			}
			h.sessions[chatID] = session
		}
	}
	return session
}

func (h *BotHandler) sendLanguageSelection(chatID int64) {
	text := "Пожалуйста, выберите язык / Тілді таңдаңыз:"
	msg := tgbotapi.NewMessage(chatID, text)

	btnRU := tgbotapi.NewKeyboardButton("🇷🇺 Русский")
	btnKK := tgbotapi.NewKeyboardButton("🇰🇿 Қазақша")
	row := []tgbotapi.KeyboardButton{btnRU, btnKK}
	keyboard := tgbotapi.NewReplyKeyboard(row)
	keyboard.ResizeKeyboard = true
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

func (h *BotHandler) sendMainMenu(chatID int64, lang string) {
	if lang == "" {
		h.sendLanguageSelection(chatID)
		return
	}

	text := "Вас приветствует Чат-бот Департамента внутреннего государственного аудита.\n\nВыберите действие:"
	if lang == "kk" {
		text = "Ішкі мемлекеттік аудит департаментінің чат-ботына қош келдіңіз.\n\nӘрекетті таңдаңыз:"
	}
	msg := tgbotapi.NewMessage(chatID, text)

	var keyboard tgbotapi.ReplyKeyboardMarkup
	if lang == "kk" {
		keyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Онлайн-қабылдау бөлмесі"),
				tgbotapi.NewKeyboardButton("Әдеп жөніндегі уәкіл"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Баспасөз орталығы"),
				tgbotapi.NewKeyboardButton("Сауалнама"),
			),
		)
	} else {
		keyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Онлайн-приемная"),
				tgbotapi.NewKeyboardButton("Уполномоченный по этике"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Пресс-центр"),
				tgbotapi.NewKeyboardButton("Анкетирование"),
			),
		)
	}
	keyboard.ResizeKeyboard = true
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

func (h *BotHandler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go h.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			go h.handleCallback(update.CallbackQuery)
		}
	}
}

func (h *BotHandler) handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := h.getSession(chatID)

	session.mu.Lock()
	defer session.mu.Unlock()

	// Обработка базовых команд
	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			session.Reset()
			session.State = StateWaitingLanguage
			h.sendLanguageSelection(chatID)
			return
		case "audit":
			h.startAuditFlow(session, chatID)
			return
		}
	}

	// Отдельная обработка контакта, так как текст при этом пустой
	if session.State == StateWaitingPhone {
		if msg.Contact != nil {
			session.PhoneNumber = msg.Contact.PhoneNumber
			session.State = StateWaitingBIN

			textMsg := "Контакт успешно подтвержден. Теперь введите БИН вашей организации (12 цифр):"
			if session.Language == "kk" {
				textMsg = "Контакт сәтті расталды. Енді ұйымыңыздың БСН-ін (12 сан) енгізіңіз:"
			}
			reply := tgbotapi.NewMessage(chatID, textMsg)
			reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			h.bot.Send(reply)
			return
		}
		
		textMsg := "Пожалуйста, используйте кнопку '📱 Отправить контакт' для продолжения."
		if session.Language == "kk" {
			textMsg = "Жалғастыру үшін '📱 Контакт жіберу' түймесін пайдаланыңыз."
		}
		reply := tgbotapi.NewMessage(chatID, textMsg)
		h.bot.Send(reply)
		return
	}

	// Для всех остальных стейтов ожидаем текстовый ввод
	text := msg.Text
	if text == "" {
		return
	}

	if session.State == StateIdle {
		switch text {
		case "Анкетирование", "Сауалнама":
			h.startAuditFlow(session, chatID)
			return
		case "Онлайн-приемная", "Онлайн-қабылдау бөлмесі":
			reply := tgbotapi.NewMessage(chatID, "Онлайн-приемная. Пожалуйста, обратитесь по телефону +7 (XXX) XXX-XX-XX.")
			if session.Language == "kk" {
				reply.Text = "Онлайн-қабылдау бөлмесі. +7 (XXX) XXX-XX-XX телефонына хабарласыңыз."
			}
			h.bot.Send(reply)
			return
		case "Уполномоченный по этике", "Әдеп жөніндегі уәкіл":
			reply := tgbotapi.NewMessage(chatID, "Уполномоченный по этике Департамента. Контактные данные: ...")
			if session.Language == "kk" {
				reply.Text = "Департаменттің Әдеп жөніндегі уәкілі. Байланыс мәліметтері: ..."
			}
			h.bot.Send(reply)
			return
		case "Пресс-центр", "Баспасөз орталығы":
			reply := tgbotapi.NewMessage(chatID, "Новости Пресс-центра: https://www.gov.kz/memleket/entities/kvga/press")
			if session.Language == "kk" {
				reply.Text = "Баспасөз орталығының жаңалықтары: https://www.gov.kz/memleket/entities/kvga/press"
			}
			h.bot.Send(reply)
			return
		}
	}

	switch session.State {
	case StateWaitingLanguage:
		if text == "🇷🇺 Русский" {
			session.Language = "ru"
		} else if text == "🇰🇿 Қазақша" {
			session.Language = "kk"
		} else {
			reply := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите язык, используя кнопки ниже / Төмендегі түймелерді пайдаланып тілді таңдаңыз:")
			h.bot.Send(reply)
			return
		}
		session.State = StateIdle
		h.sendMainMenu(chatID, session.Language)

	case StateWaitingBIN:
		if len(text) != 12 {
			msgText := "БИН должен состоять ровно из 12 символов. Попробуйте снова:"
			if session.Language == "kk" {
				msgText = "БСН дәл 12 таңбадан тұруы керек. Қайта көріңіз:"
			}
			reply := tgbotapi.NewMessage(chatID, msgText)
			h.bot.Send(reply)
			return
		}

		session.BIN = text
		session.State = StateWaitingPosition
		msgText := "Отлично. Теперь укажите вашу должность:"
		if session.Language == "kk" {
			msgText = "Өте жақсы. Енді лауазымыңызды көрсетіңіз:"
		}
		reply := tgbotapi.NewMessage(chatID, msgText)
		h.bot.Send(reply)

	case StateWaitingPosition:
		session.Position = text

		// Динамическая загрузка вопросов из БД (O(1) по индексу is_active)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		questions, err := h.repo.GetActiveQuestions(ctx)
		if err != nil || len(questions) == 0 {
			log.Printf("[ERROR] Failed to fetch active questions: %v", err)
			msgText := "Произошла ошибка при загрузке вопросов или они отсутствуют. Попробуйте позже."
			if session.Language == "kk" {
				msgText = "Сұрақтарды жүктеу кезінде қате орын алды немесе олар жоқ. Кейінірек қайталап көріңіз."
			}
			reply := tgbotapi.NewMessage(chatID, msgText)
			h.bot.Send(reply)
			session.Reset()
			h.sendMainMenu(chatID, session.Language)
			return
		}

		session.Questions = questions
		session.CurrentQIndex = 0
		session.State = StateSurveyDynamic

		h.askCurrentQuestion(session)

	case StateSurveyDynamic:
		currentQ := session.Questions[session.CurrentQIndex]
		if session.Language == "kk" {
			session.Answers[currentQ.TextKK] = text
		} else {
			session.Answers[currentQ.TextRU] = text
		}
		session.CurrentQIndex++

		if session.CurrentQIndex < len(session.Questions) {
			h.askCurrentQuestion(session)
		} else {
			session.State = StateWaitingScore

			msgText := "Анкета завершена. Оцените работу аудиторов по 5-балльной шкале (где 5 - отлично, 1 - очень плохо):"
			if session.Language == "kk" {
				msgText = "Сауалнама аяқталды. Аудиторлардың жұмысын 5 балдық жүйемен бағалаңыз (мұнда 5 - өте жақсы, 1 - өте нашар):"
			}
			reply := tgbotapi.NewMessage(chatID, msgText)
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
			msgText := "Пожалуйста, выберите число от 1 до 5 на клавиатуре."
			if session.Language == "kk" {
				msgText = "Пернетақтадан 1 мен 5 аралығындағы санды таңдаңыз."
			}
			reply := tgbotapi.NewMessage(chatID, msgText)
			h.bot.Send(reply)
			return
		}

		session.Score = score

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = h.repo.SaveAuditRecord(ctx, chatID, session.PhoneNumber, session.BIN, session.Position, session.Answers, session.Score)
		if err != nil {
			log.Printf("[ERROR] Failed to save audit record: %v", err)
			msgText := "Произошла системная ошибка при сохранении данных. Обратитесь в поддержку."
			if session.Language == "kk" {
				msgText = "Деректерді сақтау кезінде жүйелік қате орын алды. Қолдау қызметіне хабарласыңыз."
			}
			reply := tgbotapi.NewMessage(chatID, msgText)
			h.bot.Send(reply)
			session.Reset()
			h.sendMainMenu(chatID, session.Language)
			return
		}

		msgText := "Спасибо за уделенное время! Ваши ответы успешно сохранены и помогут нам улучшить качество работы."
		if session.Language == "kk" {
			msgText = "Уақыт бөлгеніңіз үшін рахмет! Жауаптарыңыз сәтті сақталды және жұмыс сапасын жақсартуға көмектеседі."
		}
		reply := tgbotapi.NewMessage(chatID, msgText)
		reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		h.bot.Send(reply)

		notification := fmt.Sprintf("Новый аудит! Телефон: %s, БИН: %s, Оценка: %d", session.PhoneNumber, session.BIN, session.Score)
		select {
		case h.leadBroker.Notifier <- []byte(notification):
		default:
			log.Println("[WARNING] SSE channel is full")
		}

		session.Reset()
		h.sendMainMenu(chatID, session.Language)

	default:
		session.Reset()
		h.sendMainMenu(chatID, session.Language)
	}
}

func (h *BotHandler) startAuditFlow(session *Session, chatID int64) {
	if session.Language == "" {
		session.State = StateWaitingLanguage
		h.sendLanguageSelection(chatID)
		return
	}
	session.Reset()
	session.State = StateWaitingPhone

	textMsg := "Для начала аудита необходимо подтвердить вашу личность. Пожалуйста, отправьте ваш номер телефона, нажав на кнопку ниже:"
	btnText := "📱 Отправить контакт"
	if session.Language == "kk" {
		textMsg = "Аудитті бастау үшін жеке басыңызды растау қажет. Төмендегі түймені басу арқылы телефон нөміріңізді жіберіңіз:"
		btnText = "📱 Контакт жіберу"
	}

	reply := tgbotapi.NewMessage(chatID, textMsg)
	btn := tgbotapi.NewKeyboardButtonContact(btnText)
	reply.ReplyMarkup = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{btn})
	h.bot.Send(reply)
}

func (h *BotHandler) askCurrentQuestion(session *Session) {
	q := session.Questions[session.CurrentQIndex]
	qText := ""
	var options []string

	if session.Language == "kk" {
		qText = fmt.Sprintf("%d. %s", session.CurrentQIndex+1, q.TextKK)
		err := json.Unmarshal(q.OptionsKK, &options)
		if err != nil {
			log.Printf("[WARNING] Не удалось распарсить варианты ответов KK для вопроса ID %d", q.ID)
			options = []string{"Иә", "Жоқ", "Қиналамын"}
		}
	} else {
		qText = fmt.Sprintf("%d. %s", session.CurrentQIndex+1, q.TextRU)
		err := json.Unmarshal(q.OptionsRU, &options)
		if err != nil {
			log.Printf("[WARNING] Не удалось распарсить варианты ответов RU для вопроса ID %d", q.ID)
			options = []string{"Да", "Нет", "Затрудняюсь"}
		}
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