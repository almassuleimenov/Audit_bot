package bot

//D:\Project\backend_projects\audit_bot\bot\fsm.go
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/almassuleimenov/Audit_bot/internal/sse"
	"github.com/almassuleimenov/Audit_bot/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	StateLanguage = iota
	StateMenu
	StateAuditWarning
	StateAuditPosition
	StateAuditBin
	StateAuditQuestions
	StateAuditScore
	StateAppManager
	StateAppFIO
	StateAppQuestion
	StateAppPhone
)

// Компилируем регулярное выражение один раз при старте (O(1) в рантайме)
var binRegex = regexp.MustCompile(`^\d{12}$`)

type UserState struct {
	mu           sync.Mutex
	LastActivity time.Time
	Step         int
	Language     string

	Position      string
	BIN           string
	Answers       map[string]string
	QuestionIndex int
	Questions     []repository.SurveyQuestion

	TargetManager string
	FullName      string
	Question      string
	PhoneNumber   string
}

var translations = map[string]map[string]string{
	"ru": {
		"menu_greet":     "Вас приветствует Чат-бот Департамента внутреннего государственного аудита. Выберите раздел:",
		"btn_audit":      "Анкетирование",
		"btn_app":        "Онлайн-приемная",
		"btn_ethics":     "Уполномоченный по этике",
		"btn_press":      "Пресс-центр",
		"press_news":     "Новости",
		"press_schedule": "График",
		"press_title":    "Пресс-центр Комитета:",
		"ethics_title":   "Уполномоченный по этике Департамента",
		"audit_warning":  "Анкета предназначена для мониторинга соблюдения сотрудниками... Заведомо ложные ответы влекут ответственность.",
		"btn_agree":      "Ознакомлен ✅",
		"btn_disagree":   "Не ознакомлен ❌",
		"ask_position":   "Пожалуйста, укажите Вашу должность для продолжения:",
		"ask_bin":        "Пожалуйста, укажите Ваш БИН (12 цифр):",
		"err_bin":        "Некорректный БИН. Введите ровно 12 цифр:",
		"ask_score":      "Благодарим за участие. Оцените аудитора от 1 до 5:",
		"err_score":      "Пожалуйста, введите число от 1 до 5.",
		"err_save":       "Произошла ошибка при сохранении. Попробуйте позже.",
		"audit_success":  "Принято ✅. Ваши ответы сохранены.",
		"app_managers":   "Выберите руководителя для записи:",
		"ask_fio":        "Укажите ваше ФИО:",
		"ask_question":   "Характер вопроса кратко:",
		"ask_phone":      "Введите номер телефона для обратной связи:",
		"app_success":    "Вы успешно записались на прием! С Вами свяжутся в ближайшее время.",
		"err_files":      "Бот не обрабатывает файлы. Пожалуйста, отправьте текст.",
		"err_lang":       "Пожалуйста, выберите язык, нажав на кнопку ниже.",
	},
	"kk": {
		"menu_greet":     "Ішкі мемлекеттік аудит департаментінің чат-боты сізді қарсы алады. Бөлімді таңдаңыз:",
		"btn_audit":      "Сауалнама",
		"btn_app":        "Онлайн-қабылдау",
		"btn_ethics":     "Әдеп жөніндегі уәкіл",
		"btn_press":      "Баспасөз орталығы",
		"press_news":     "Жаңалықтар",
		"press_schedule": "Кесте",
		"press_title":    "Комитеттің баспасөз орталығы:",
		"ethics_title":   "Департаменттің әдеп жөніндегі уәкілі",
		"audit_warning":  "Сауалнама қызметкерлердің сақталуын бақылауға арналған... Көрінеу жалған жауаптар жауапкершілікке әкеп соғады.",
		"btn_agree":      "Таныстым ✅",
		"btn_disagree":   "Таныспадым ❌",
		"ask_position":   "Жалғастыру үшін лауазымыңызды көрсетіңіз:",
		"ask_bin":        "БСН көрсетіңіз (12 сан):",
		"err_bin":        "Қате БСН. Толық 12 сан енгізіңіз:",
		"ask_score":      "Қатысқаныңыз үшін рахмет. Аудиторды 1-ден 5-ке дейін бағалаңыз:",
		"err_score":      "1-ден 5-ке дейін сан енгізіңіз.",
		"err_save":       "Сақтау кезінде қате пайда болды. Кейінірек қайталап көріңіз.",
		"audit_success":  "Қабылданды ✅. Сіздің жауаптарыңыз сақталды.",
		"app_managers":   "Жазылу үшін басшыны таңдаңыз:",
		"ask_fio":        "Аты-жөніңізді көрсетіңіз (ТАӘ):",
		"ask_question":   "Сұрақтың мәнін қысқаша жазыңыз:",
		"ask_phone":      "Кері байланыс үшін телефон нөмірін енгізіңіз:",
		"app_success":    "Сіз қабылдауға сәтті жазылдыңыз! Сізбен жақын арада байланысады.",
		"err_files":      "Бот файлдарды өңдемейді. Мәтін жіберіңіз.",
		"err_lang":       "Төмендегі түймені басу арқылы тілді таңдаңыз.",
	},
}

type BotHandler struct {
	bot         *tgbotapi.BotAPI
	repo        repository.BotRepository
	sessions    sync.Map
	adminChatID int64
	broker      *sse.Broker // Интеграция SSE
}

func NewBotHandler(bot *tgbotapi.BotAPI, repo repository.BotRepository, adminID int64, broker *sse.Broker) *BotHandler {
	return &BotHandler{
		bot:         bot,
		repo:        repo,
		adminChatID: adminID,
		broker:      broker,
	}
}

func (h *BotHandler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := h.bot.GetUpdatesChan(u)

	go h.startSessionCleaner(1 * time.Hour)

	for update := range updates {
		go h.processUpdate(update)
	}
}

func (h *BotHandler) startSessionCleaner(ttl time.Duration) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		h.sessions.Range(func(key, value interface{}) bool {
			state := value.(*UserState)
			state.mu.Lock()
			lastActivity := state.LastActivity
			state.mu.Unlock()
			if now.Sub(lastActivity) > ttl {
				h.sessions.Delete(key)
			}
			return true
		})
	}
}

func (h *BotHandler) getOrCreateState(chatID int64) *UserState {
	val, loaded := h.sessions.LoadOrStore(chatID, &UserState{
		Step:         StateLanguage,
		Answers:      make(map[string]string),
		LastActivity: time.Now(),
	})
	state := val.(*UserState)
	if !loaded {
		state.Answers = make(map[string]string)
	}
	return state
}

func (h *BotHandler) processUpdate(update tgbotapi.Update) {
	var chatID int64
	var text, callbackData string

	if update.Message != nil {
		chatID = update.Message.Chat.ID
		text = update.Message.Text
		if update.Message.Photo != nil || update.Message.Document != nil {
			state := h.getOrCreateState(chatID)
			lang := state.Language
			if lang == "" {
				lang = "ru"
			}
			h.sendMessage(chatID, translations[lang]["err_files"])
			return
		}
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callbackData = update.CallbackQuery.Data
		h.bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
	} else {
		return
	}

	state := h.getOrCreateState(chatID)

	// Внимание: Блокируем мьютекс только на обновление активности.
	// Обертывание долгих I/O операций (DB, API) в мьютекс — антипаттерн SRE,
	// так как это может заблокировать другие горутины.
	state.mu.Lock()
	state.LastActivity = time.Now()
	state.mu.Unlock()

	if text == "/start" {
		state.Step = StateLanguage
		state.Language = ""
		state.Answers = make(map[string]string)
		h.sendLanguageSelection(chatID)
		return
	}

	switch state.Step {
	case StateLanguage:
		if callbackData == "lang_ru" || callbackData == "lang_kk" {
			state.Language = callbackData[5:]
			state.Step = StateMenu
			h.sendMenu(chatID, state.Language)
		} else {
			h.sendLanguageSelection(chatID)
		}

	case StateMenu:
		h.handleMenuChoice(chatID, callbackData, state)

	case StateAuditWarning:
		if callbackData == "audit_agree" {
			state.Step = StateAuditPosition
			h.sendMessage(chatID, translations[state.Language]["ask_position"])
		} else if callbackData == "audit_disagree" {
			state.Step = StateMenu
			h.sendMenu(chatID, state.Language)
		}

	case StateAuditPosition:
		if text != "" {
			state.Position = text
			state.Step = StateAuditBin
			h.sendMessage(chatID, translations[state.Language]["ask_bin"])
		}

	case StateAuditBin:
		if h.isValidBIN(text) {
			state.BIN = text

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			questions, err := h.repo.GetActiveQuestions(ctx)
			if err != nil || len(questions) == 0 {
				h.sendMessage(chatID, "В данный момент анкета недоступна. Попробуйте позже.")
				state.Step = StateMenu
				return
			}
			state.Questions = questions
			state.QuestionIndex = 0
			state.Step = StateAuditQuestions
			h.sendDynamicAuditQuestion(chatID, state)
		} else {
			h.sendMessage(chatID, translations[state.Language]["err_bin"])
		}

	case StateAuditQuestions:
		if callbackData != "" {
			currentQuestion := state.Questions[state.QuestionIndex]
			qKey := fmt.Sprintf("question_%d", currentQuestion.ID)

			state.Answers[qKey] = callbackData
			state.QuestionIndex++

			if state.QuestionIndex < len(state.Questions) {
				h.sendDynamicAuditQuestion(chatID, state)
			} else {
				state.Step = StateAuditScore
				h.sendMessage(chatID, translations[state.Language]["ask_score"])
			}
		}

	case StateAuditScore:
		score, err := strconv.Atoi(text)
		if err != nil || score < 1 || score > 5 {
			h.sendMessage(chatID, translations[state.Language]["err_score"])
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = h.repo.SaveAuditRecord(ctx, chatID, state.BIN, state.Position, state.Answers, score)
		if err == nil {
			h.sendMessage(chatID, translations[state.Language]["audit_success"])
		}
		state.Step = StateMenu
		h.sendMenu(chatID, state.Language)

	case StateAppManager:
		if callbackData != "" {
			state.TargetManager = callbackData
			state.Step = StateAppFIO
			h.sendMessage(chatID, translations[state.Language]["ask_fio"])
		}
	case StateAppFIO:
		if text != "" {
			state.FullName = text
			state.Step = StateAppQuestion
			h.sendMessage(chatID, translations[state.Language]["ask_question"])
		}
	case StateAppQuestion:
		if text != "" {
			state.Question = text
			state.Step = StateAppPhone
			h.sendMessage(chatID, translations[state.Language]["ask_phone"])
		}
	case StateAppPhone:
		if text != "" {
			state.PhoneNumber = text
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Сохраняем в БД
			h.repo.SaveAppointment(ctx, chatID, state.TargetManager, state.FullName, state.PhoneNumber, state.Question)

			// 1. Отправляем уведомление админу
			adminMsg := fmt.Sprintf("🔔 *Новая запись на прием!*\n\n*К кому:* %s\n*ФИО:* %s\n*Вопрос:* %s\n*Телефон:* %s",
				state.TargetManager, state.FullName, state.Question, state.PhoneNumber)
			msg := tgbotapi.NewMessage(h.adminChatID, adminMsg)
			msg.ParseMode = "Markdown"
			h.bot.Send(msg)

			// 2. [НОВОЕ] Формируем и отправляем событие в SSE Брокер для дашборда (O(1))
			// Генерируем уникальный ID для события (chatID + timestamp)
			leadEvent := map[string]interface{}{
				"id":       fmt.Sprintf("%d-%d", chatID, time.Now().UnixNano()),
				"phone":    state.PhoneNumber,
				"username": state.FullName,
				"status":   "NEW",
			}

			if eventBytes, err := json.Marshal(leadEvent); err == nil {
				// Асинхронно пушим в канал, чтобы не блокировать бота, если брокер перегружен
				select {
				case h.broker.Notifier <- eventBytes:
				default:
					log.Println("[WARNING] SSE Broker is full or blocked. Dropped event.")
				}
			}

			h.sendMessage(chatID, translations[state.Language]["app_success"])
			h.sessions.Delete(chatID)
			state.Step = StateMenu
			h.sendMenu(chatID, state.Language)
		}
	}
}

func (h *BotHandler) sendDynamicAuditQuestion(chatID int64, state *UserState) {
	lang := state.Language
	index := state.QuestionIndex
	question := state.Questions[index]

	var qText string
	var optionsBytes []byte

	if lang == "ru" {
		qText = question.TextRU
		optionsBytes = question.OptionsRU
	} else {
		qText = question.TextKK
		optionsBytes = question.OptionsKK
	}

	var options []string
	// Добавлена проверка на ошибку парсинга. Без нее битый JSON в базе положит бота.
	if err := json.Unmarshal(optionsBytes, &options); err != nil {
		log.Printf("[ERROR] Failed to unmarshal options for question %d: %v", question.ID, err)
		options = []string{"Да", "Нет"} // Fallback-вариант
	}

	msg := tgbotapi.NewMessage(chatID, qText)

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, opt := range options {
		btn := tgbotapi.NewInlineKeyboardButtonData(opt, opt)
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(btn))
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	h.bot.Send(msg)
}

func (h *BotHandler) sendLanguageSelection(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "🇷🇺 Выберите язык интерфейса\n🇰🇿 Интерфейс тілін таңдаңыз")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_ru"),
			tgbotapi.NewInlineKeyboardButtonData("🇰🇿 Қазақша", "lang_kk"),
		),
	)
	h.bot.Send(msg)
}

func (h *BotHandler) sendMessage(chatID int64, text string) {
	h.bot.Send(tgbotapi.NewMessage(chatID, text))
}

func (h *BotHandler) sendDocument(chatID int64, fileURL string) {
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileURL(fileURL))
	h.bot.Send(doc)
}

func (h *BotHandler) sendMenu(chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, translations[lang]["menu_greet"])
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(translations[lang]["btn_audit"], "menu_audit")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(translations[lang]["btn_app"], "menu_app")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(translations[lang]["btn_ethics"], "menu_ethics")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(translations[lang]["btn_press"], "menu_press")),
	)
	h.bot.Send(msg)
}

func (h *BotHandler) handleMenuChoice(chatID int64, data string, state *UserState) {
	lang := state.Language
	switch data {
	case "menu_press":
		msg := tgbotapi.NewMessage(chatID, translations[lang]["press_title"])
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(translations[lang]["press_news"], "https://www.gov.kz/memleket/entities/kvga/press?lang=ru")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(translations[lang]["press_schedule"], "https://www.gov.kz/memleket/entities/kvga/about/structure/departments/activity/4728/1?lang=ru")),
		)
		h.bot.Send(msg)
		h.sendMenu(chatID, lang)
	case "menu_ethics":
		h.sendDocument(chatID, "https://robochat.storage.yandexcloud.net/attachments/day/20284/421499/file/OLnAb2ZL/%D3%98%D0%B4%D0%B5%D0%BF%20%D0%B3%D1%80%D0%B0%D1%84%D0%B8%D0%BA.pdf")
		h.sendMessage(chatID, translations[lang]["ethics_title"])
		h.sendMenu(chatID, lang)
	case "menu_audit":
		state.Step = StateAuditWarning
		h.sendDocument(chatID, "https://robochat.storage.yandexcloud.net/attachments/day/20285/421499/file/gawYYoV4/%D0%9C%D0%B5%D1%82%D0%BE%D0%B4%D0%B8%D1%87%D0%BA%D0%B0%20%D0%B4%D0%BB%D1%8F%20%D1%81%D0%BE%D1%82%D1%80%20%281%29.pdf")
		msg := tgbotapi.NewMessage(chatID, translations[lang]["audit_warning"])
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(translations[lang]["btn_agree"], "audit_agree"),
				tgbotapi.NewInlineKeyboardButtonData(translations[lang]["btn_disagree"], "audit_disagree"),
			),
		)
		h.bot.Send(msg)
	case "menu_app":
		state.Step = StateAppManager
		msg := tgbotapi.NewMessage(chatID, translations[lang]["app_managers"])
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Қабдыраш Б.С.", "manager_kabdrash")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Мұсабек А.М.", "manager_musabek")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Джумагулов М.Б.", "manager_dzhumagulov")),
		)
		h.bot.Send(msg)
	}
}

func (h *BotHandler) isValidBIN(bin string) bool {
	return binRegex.MatchString(bin)
}
