package bot

// D:\Project\backend_projects\audit_bot\bot\fsm.go

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/almassuleimenov/Audit_bot/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Состояния FSM
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

// UserState хранит текущий контекст диалога пользователя
type UserState struct {
	Step          int
	Language      string // "ru" или "kz"
	Position      string
	BIN           string
	Answers       map[string]string
	QuestionIndex int
	TargetManager string
	FullName      string
	Question      string
	PhoneNumber   string
}

// I18nText описывает структуру всех текстовых данных в боте
type I18nText struct {
	Welcome      string
	MenuAudit    string
	MenuApp      string
	MenuEthics   string
	MenuPress    string
	FileError    string
	PressCenter  string
	News         string
	Schedule     string
	EthicsInfo   string
	AuditWarning string
	BtnAgree     string
	BtnDisagree  string
	AskPosition  string
	AskBIN       string
	InvalidBIN   string
	BtnYes       string
	BtnNo        string
	BtnDunno     string
	AskScore     string
	InvalidScore string
	SaveError    string
	SaveSuccess  string
	AskManager   string
	Manager1     string
	Manager2     string
	Manager3     string
	AskFIO       string
	AskQuestion  string
	AskPhone     string
	AppSuccess   string
	Questions    []string
}

// Хэш-таблица со всеми переводами (O(1) доступ)
var dict = map[string]I18nText{
	"ru": {
		Welcome:      "Вас приветствует Чат-бот Департамента внутреннего государственного аудита. Выберите раздел:",
		MenuAudit:    "Анкетирование",
		MenuApp:      "Онлайн-приемная",
		MenuEthics:   "Уполномоченный по этике",
		MenuPress:    "Пресс-центр",
		FileError:    "Бот не обрабатывает файлы. Пожалуйста, оставьте контактные данные.",
		PressCenter:  "Пресс-центр Комитета:",
		News:         "Новости",
		Schedule:     "График",
		EthicsInfo:   "Уполномоченный по этике Департамента",
		AuditWarning: "Анкета предназначена для мониторинга соблюдения сотрудниками внутреннего государственного аудита антикоррупционного законодательства. Заведомо ложные ответы влекут ответственность (ст. 419, 274 УК РК).",
		BtnAgree:     "Ознакомлен ✅",
		BtnDisagree:  "Не ознакомлен ❌",
		AskPosition:  "Пожалуйста, укажите Вашу должность для продолжения:",
		AskBIN:       "Пожалуйста, укажите Ваш БИН (12 цифр):",
		InvalidBIN:   "Некорректный БИН. Введите ровно 12 цифр:",
		BtnYes:       "Да",
		BtnNo:        "Нет",
		BtnDunno:     "Затрудняюсь ответить",
		AskScore:     "Благодарим за участие. Оцените аудитора от 1 до 5:",
		InvalidScore: "Пожалуйста, введите число от 1 до 5.",
		SaveError:    "Произошла ошибка при сохранении. Попробуйте позже.",
		SaveSuccess:  "Принято ✅. Ваши ответы сохранены.",
		AskManager:   "Выберите руководителя для записи:",
		Manager1:     "Қабдыраш Б.С.",
		Manager2:     "Мұсабек А.М.",
		Manager3:     "Джумагулов М.Б.",
		AskFIO:       "Укажите ваше ФИО:",
		AskQuestion:  "Характер вопроса кратко:",
		AskPhone:     "Введите номер телефона для обратной связи:",
		AppSuccess:   "Вы успешно записались на прием! С Вами свяжутся в ближайшее время.",
		Questions: []string{
			"1. Были ли сотрудники аудита вежливы, корректны и уважительны в общении в процессе проверки?",
			"2. Возникали ли в процессе проверки ситуации, которые, по вашему мнению, могли носить признаки давления или нарушения этических норм со стороны аудиторов?",
			"3. Предлагал ли кто-либо из сотрудников аудита неофициальное или \"альтернативное\" решение по вопросам, связанным с результатами проверки?",
			"4. Были ли попытки получения личной выгоды (подарки, деньги, услуги и т. д.) со стороны проверяющих?",
			"5. Демонстрировали ли аудиторы прозрачность и объективность в оценке проверяемой информации?",
			"6. Вмешивались ли аудиторы в деятельность организации за рамками своей компетенции?",
			"7. Придерживались ли сотрудники аудита принципов конфиденциальности при обращении с внутренней документацией и информацией?",
			"8. Возникали ли у вас ощущения, что аудиторы предвзято относятся к вашей организации или отдельным сотрудникам?",
			"9. Оказывалось ли к вам или вашим коллегам давление с целью изменить или скрыть какую-либо информацию в ходе аудита?",
		},
	},
	"kz": {
		Welcome:      "Ішкі мемлекеттік аудит департаментінің чат-ботына қош келдіңіз. Бөлімді таңдаңыз:",
		MenuAudit:    "Сауалнама",
		MenuApp:      "Онлайн-қабылдау",
		MenuEthics:   "Әдеп жөніндегі уәкіл",
		MenuPress:    "Баспасөз орталығы",
		FileError:    "Бот файлдарды қабылдамайды. Байланыс деректеріңізді қалдырыңыз.",
		PressCenter:  "Комитеттің баспасөз орталығы:",
		News:         "Жаңалықтар",
		Schedule:     "Кесте",
		EthicsInfo:   "Департаменттің әдеп жөніндегі уәкілі",
		AuditWarning: "Сауалнама қызметкерлердің сыбайлас жемқорлыққа қарсы заңнаманы сақтауын мониторингтеуге арналған. Көрінеу жалған жауаптар жауапкершілікке әкеп соғады (ҚР ҚК 419, 274-баптары).",
		BtnAgree:     "Таныстым ✅",
		BtnDisagree:  "Таныспадым ❌",
		AskPosition:  "Жалғастыру үшін лауазымыңызды көрсетіңіз:",
		AskBIN:       "БСН көрсетіңіз (12 сан):",
		InvalidBIN:   "БСН қате. Толық 12 сан енгізіңіз:",
		BtnYes:       "Иә",
		BtnNo:        "Жоқ",
		BtnDunno:     "Жауап беруге қиналамын",
		AskScore:     "Қатысқаныңыз үшін рахмет. Аудиторды 1-ден 5-ке дейін бағалаңыз:",
		InvalidScore: "1-ден 5-ке дейінгі санды енгізіңіз.",
		SaveError:    "Сақтау кезінде қате пайда болды. Кейінірек қайталап көріңіз.",
		SaveSuccess:  "Қабылданды ✅. Сіздің жауаптарыңыз сақталды.",
		AskManager:   "Жазылу үшін басшыны таңдаңыз:",
		Manager1:     "Қабдыраш Б.С.",
		Manager2:     "Мұсабек А.М.",
		Manager3:     "Жұмағұлов М.Б.",
		AskFIO:       "Аты-жөніңізді (ТіАӘ) көрсетіңіз:",
		AskQuestion:  "Сұрақтың мәні қысқаша:",
		AskPhone:     "Кері байланыс үшін телефон нөмірін енгізіңіз:",
		AppSuccess:   "Сіз қабылдауға сәтті жазылдыңыз! Сізбен жақын арада хабарласамыз.",
		Questions: []string{
			"1. Тексеру барысында аудит қызметкерлері сыпайы, әдепті және құрметпен сөйлесті ме?",
			"2. Тексеру барысында аудиторлар тарапынан қысым көрсету немесе әдеп нормаларын бұзу белгілері болуы мүмкін жағдайлар туындады ма?",
			"3. Аудит қызметкерлерінің біреуі тексеру нәтижелеріне байланысты бейресми немесе \"балама\" шешім ұсынды ма?",
			"4. Тексерушілер тарапынан жеке пайда алу (сыйлықтар, ақша, қызметтер және т.б.) әрекеттері болды ма?",
			"5. Аудиторлар тексерілетін ақпаратты бағалауда ашықтық пен объективтілікті көрсетті ме?",
			"6. Аудиторлар өз құзыретінен тыс ұйымның қызметіне араласты ма?",
			"7. Аудит қызметкерлері ішкі құжаттама мен ақпаратты пайдалану кезінде құпиялылық принциптерін ұстанды ма?",
			"8. Сізде аудиторлар сіздің ұйымыңызға немесе жекелеген қызметкерлерге біржақты қарайды деген сезім туындады ма?",
			"9. Аудит барысында қандай да бір ақпаратты өзгерту немесе жасыру мақсатында сізге немесе сіздің әріптестеріңізге қысым көрсетілді ме?",
		},
	},
}

// Вспомогательная функция для безопасного получения нужного словаря
func getText(lang string) I18nText {
	if lang == "" {
		lang = "ru"
	}
	if t, ok := dict[lang]; ok {
		return t
	}
	return dict["ru"]
}

type BotHandler struct {
	bot      *tgbotapi.BotAPI
	repo     repository.BotRepository
	sessions sync.Map
}

func NewBotHandler(bot *tgbotapi.BotAPI, repo repository.BotRepository) *BotHandler {
	return &BotHandler{
		bot:  bot,
		repo: repo,
	}
}

func (h *BotHandler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := h.bot.GetUpdatesChan(u)
	log.Println("[INFO] FSM Engine started")

	for update := range updates {
		go h.processUpdate(update)
	}
}

func (h *BotHandler) processUpdate(update tgbotapi.Update) {
	var chatID int64
	var text, callbackData string
	var messageID int

	if update.Message != nil {
		chatID = update.Message.Chat.ID
		text = update.Message.Text
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callbackData = update.CallbackQuery.Data
		messageID = update.CallbackQuery.Message.MessageID
		h.bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
	} else {
		return
	}

	// Получаем или создаем состояние юзера
	val, ok := h.sessions.Load(chatID)
	var state UserState
	if ok {
		state = val.(UserState)
	} else {
		state = UserState{Step: StateLanguage, Answers: make(map[string]string)}
	}

	t := getText(state.Language)

	// Защита от медиафайлов
	if update.Message != nil && (update.Message.Photo != nil || update.Message.Document != nil || update.Message.Video != nil || update.Message.Audio != nil) {
		h.sendMessage(chatID, t.FileError)
		return
	}

	// Команда сброса состояния (перехватываем /start)
	if text == "/start" {
		state = UserState{Step: StateLanguage, Answers: make(map[string]string)}
		h.sendLanguageSelection(chatID)
		h.sessions.Store(chatID, state)
		return
	}

	// Роутер состояний
	switch state.Step {
	case StateLanguage:
		if callbackData == "lang_ru" || callbackData == "lang_kz" {
			state.Language = callbackData[5:] // Извлекаем ru или kz
			state.Step = StateMenu
			h.sendMenu(chatID, state.Language)
		} else {
			h.sendLanguageSelection(chatID)
		}

	case StateMenu:
		h.handleMenuChoice(chatID, callbackData, &state)

	case StateAuditWarning:
		h.handleAuditWarning(chatID, callbackData, messageID, &state)

	case StateAuditPosition:
		if text != "" {
			state.Position = text
			state.Step = StateAuditBin
			h.sendMessage(chatID, t.AskBIN)
		}

	case StateAuditBin:
		if h.isValidBIN(text) {
			state.BIN = text
			state.Step = StateAuditQuestions
			state.QuestionIndex = 0
			h.sendAuditQuestion(chatID, &state)
		} else {
			h.sendMessage(chatID, t.InvalidBIN)
		}

	case StateAuditQuestions:
		if callbackData != "" {
			qKey := fmt.Sprintf("q%d", state.QuestionIndex+1)
			state.Answers[qKey] = callbackData
			state.QuestionIndex++

			if state.QuestionIndex < len(t.Questions) {
				h.sendAuditQuestion(chatID, &state)
			} else {
				state.Step = StateAuditScore
				h.sendMessage(chatID, t.AskScore)
			}
		}

	case StateAuditScore:
		score, err := strconv.Atoi(text)
		if err != nil || score < 1 || score > 5 {
			h.sendMessage(chatID, t.InvalidScore)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = h.repo.SaveAuditRecord(ctx, chatID, state.BIN, state.Position, state.Answers, score)
		if err != nil {
			log.Printf("[ERROR] SaveAuditRecord: %v", err)
			h.sendMessage(chatID, t.SaveError)
		} else {
			h.sendMessage(chatID, t.SaveSuccess)
		}

		state.Step = StateMenu
		h.sendMenu(chatID, state.Language)

	case StateAppManager:
		if callbackData != "" {
			state.TargetManager = callbackData
			state.Step = StateAppFIO
			h.sendMessage(chatID, t.AskFIO)
		}

	case StateAppFIO:
		if text != "" {
			state.FullName = text
			state.Step = StateAppQuestion
			h.sendMessage(chatID, t.AskQuestion)
		}

	case StateAppQuestion:
		if text != "" {
			state.Question = text
			state.Step = StateAppPhone
			h.sendMessage(chatID, t.AskPhone)
		}

	case StateAppPhone:
		if text != "" {
			state.PhoneNumber = text

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := h.repo.SaveAppointment(ctx, chatID, state.TargetManager, state.FullName, state.PhoneNumber, state.Question)
			if err != nil {
				log.Printf("[ERROR] SaveAppointment: %v", err)
				h.sendMessage(chatID, t.SaveError)
			} else {
				adminChatID := int64(601610)
				adminMsg := fmt.Sprintf("🔔 *Новая запись на прием!*\n\n*К кому:* %s\n*ФИО:* %s\n*Вопрос:* %s\n*Телефон:* %s",
					state.TargetManager, state.FullName, state.Question, state.PhoneNumber)

				msg := tgbotapi.NewMessage(adminChatID, adminMsg)
				msg.ParseMode = "Markdown"

				if _, err := h.bot.Send(msg); err != nil {
					log.Printf("[ERROR] Не удалось отправить алерт админам: %v", err)
				}

				h.sendMessage(chatID, t.AppSuccess)
			}
			state.Step = StateMenu
			h.sendMenu(chatID, state.Language)
		}
	}

	h.sessions.Store(chatID, state)
}

// --- Вспомогательные методы ---

func (h *BotHandler) sendMessage(chatID int64, text string) {
	if text == "" {
		log.Printf("[WARNING] Attempted to send empty message to %d", chatID)
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("[ERROR] Failed to send message: %v", err)
	}
}

func (h *BotHandler) sendDocument(chatID int64, fileURL string) {
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileURL(fileURL))
	if _, err := h.bot.Send(doc); err != nil {
		log.Printf("[ERROR] Failed to send document to %d: %v", chatID, err)
	}
}

// Выбор языка
func (h *BotHandler) sendLanguageSelection(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите язык / Тілді таңдаңыз:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_ru"),
			tgbotapi.NewInlineKeyboardButtonData("🇰🇿 Қазақша", "lang_kz"),
		),
	)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("[ERROR] Failed to send language selection: %v", err)
	}
}

// Генерация главного меню
func (h *BotHandler) sendMenu(chatID int64, lang string) {
	t := getText(lang)
	msg := tgbotapi.NewMessage(chatID, t.Welcome)

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(t.MenuAudit, "menu_audit")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(t.MenuApp, "menu_app")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(t.MenuEthics, "menu_ethics")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(t.MenuPress, "menu_press")),
	)

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("[ERROR] Failed to send menu: %v", err)
	}
}

func (h *BotHandler) handleMenuChoice(chatID int64, data string, state *UserState) {
	t := getText(state.Language)

	switch data {
	case "menu_press":
		msg := tgbotapi.NewMessage(chatID, t.PressCenter)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(t.News, "https://www.gov.kz/memleket/entities/kvga/press?lang=ru")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(t.Schedule, "https://www.gov.kz/memleket/entities/kvga/about/structure/departments/activity/4728/1?lang=ru")),
		)
		if _, err := h.bot.Send(msg); err != nil {
			log.Printf("[ERROR] Failed to send press center: %v", err)
		}
		h.sendMenu(chatID, state.Language)

	case "menu_ethics":
		h.sendDocument(chatID, "https://robochat.storage.yandexcloud.net/attachments/day/20284/421499/file/OLnAb2ZL/%D3%98%D0%B4%D0%B5%D0%BF%20%D0%B3%D1%80%D0%B0%D1%84%D0%B8%D0%BA.pdf")
		h.sendMessage(chatID, t.EthicsInfo)
		h.sendMenu(chatID, state.Language)

	case "menu_audit":
		state.Step = StateAuditWarning
		h.sendDocument(chatID, "https://robochat.storage.yandexcloud.net/attachments/day/20285/421499/file/gawYYoV4/%D0%9C%D0%B5%D1%82%D0%BE%D0%B4%D0%B8%D1%87%D0%BA%D0%B0%20%D0%B4%D0%BB%D1%8F%20%D1%81%D0%BE%D1%82%D1%80%20%281%29.pdf")

		msg := tgbotapi.NewMessage(chatID, t.AuditWarning)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(t.BtnAgree, "audit_agree"),
				tgbotapi.NewInlineKeyboardButtonData(t.BtnDisagree, "audit_disagree"),
			),
		)
		if _, err := h.bot.Send(msg); err != nil {
			log.Printf("[ERROR] Failed to send audit warning: %v", err)
		}

	case "menu_app":
		state.Step = StateAppManager
		msg := tgbotapi.NewMessage(chatID, t.AskManager)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(t.Manager1, "manager_kabdrash")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(t.Manager2, "manager_musabek")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(t.Manager3, "manager_dzhumagulov")),
		)
		if _, err := h.bot.Send(msg); err != nil {
			log.Printf("[ERROR] Failed to send manager list: %v", err)
		}
	}
}

func (h *BotHandler) handleAuditWarning(chatID int64, data string, messageID int, state *UserState) {
	t := getText(state.Language)

	if data == "audit_agree" {
		state.Step = StateAuditPosition
		h.sendMessage(chatID, t.AskPosition)
	} else if data == "audit_disagree" {
		h.sendMenu(chatID, state.Language)
	}
}

func (h *BotHandler) sendAuditQuestion(chatID int64, state *UserState) {
	t := getText(state.Language)
	msg := tgbotapi.NewMessage(chatID, t.Questions[state.QuestionIndex])

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(t.BtnYes, t.BtnYes),
			tgbotapi.NewInlineKeyboardButtonData(t.BtnNo, t.BtnNo),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(t.BtnDunno, t.BtnDunno),
		),
	)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("[ERROR] Failed to send audit question: %v", err)
	}
}

func (h *BotHandler) isValidBIN(bin string) bool {
	match, _ := regexp.MatchString(`^\d{12}$`, bin)
	return match
}