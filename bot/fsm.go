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
// Добавлен sync.Mutex для безопасной конкурентной работы
type UserState struct {
	mu            sync.Mutex
	Step          int
	Language      string // "ru" или "kk"
	
	// Данные для Аудита
	Position      string
	BIN           string
	Answers       map[string]string
	QuestionIndex int
	
	// Данные для Приемной
	TargetManager string
	FullName      string
	Question      string
	PhoneNumber   string
}

// i18n словари локализации
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
		"audit_warning":  "Анкета предназначена для мониторинга соблюдения сотрудниками... Заведомо ложные ответы влекут ответственность (ст. 419, 274 УК РК).",
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
		"btn_yes":        "Да",
		"btn_no":         "Нет",
		"btn_idk":        "Затрудняюсь ответить",
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
		"audit_warning":  "Сауалнама қызметкерлердің сақталуын бақылауға арналған... Көрінеу жалған жауаптар жауапкершілікке әкеп соғады (ҚР ҚК 419, 274 баптары).",
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
		"btn_yes":        "Иә",
		"btn_no":         "Жоқ",
		"btn_idk":        "Жауап беруге қиналамын",
		"err_lang":       "Төмендегі түймені басу арқылы тілді таңдаңыз.",
	},
}

var auditQuestions = map[string][]string{
	"ru": {
		"1. Были ли сотрудники аудита вежливы, корректны и уважительны в общении в процессе проверки?",
		"2. Возникали ли в процессе проверки ситуации, которые могли носить признаки давления?",
		"3. Предлагал ли кто-либо из сотрудников неофициальное решение по результатам проверки?",
		"4. Были ли попытки получения личной выгоды со стороны проверяющих?",
		"5. Демонстрировали ли аудиторы прозрачность и объективность?",
		"6. Вмешивались ли аудиторы в деятельность организации за рамками компетенции?",
		"7. Придерживались ли сотрудники аудита принципов конфиденциальности?",
		"8. Возникали ли у вас ощущения предвзятого отношения?",
		"9. Оказывалось ли к вам давление с целью скрыть информацию?",
	},
	"kk": {
		"1. Аудит қызметкерлері тексеру барысында сыпайы, әдепті сөйлесті ме?",
		"2. Тексеру барысында аудиторлар тарапынан қысым көрсету белгілері байқалды ма?",
		"3. Аудит қызметкерлері бейресми шешім ұсынды ма?",
		"4. Тексерушілер тарапынан жеке пайда алу әрекеттері болды ма?",
		"5. Аудиторлар объективтілікті көрсетті ме?",
		"6. Аудиторлар өз құзыретінен тыс ұйымның қызметіне араласты ма?",
		"7. Аудит қызметкерлері құпиялылық қағидаттарын сақтады ма?",
		"8. Аудиторлар сіздің ұйымыңызға біржақты қарайды деген сезім болды ма?",
		"9. Аудит барысында ақпаратты жасыру мақсатында сізге қысым көрсетілді ме?",
	},
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

	log.Println("[INFO] FSM Engine started (Multi-language, Thread-safe)")

	for update := range updates {
		go h.processUpdate(update)
	}
}

// getOrCreateState атомарно извлекает или создает состояние пользователя
func (h *BotHandler) getOrCreateState(chatID int64) *UserState {
	val, loaded := h.sessions.LoadOrStore(chatID, &UserState{
		Step:    StateLanguage,
		Answers: make(map[string]string),
	})
	
	state := val.(*UserState)
	
	// Если это новая сессия, гарантируем инициализацию карты
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

		// Защита от медиафайлов
		if update.Message.Photo != nil || update.Message.Document != nil || update.Message.Video != nil || update.Message.Audio != nil {
			state := h.getOrCreateState(chatID)
			state.mu.Lock()
			lang := state.Language
			state.mu.Unlock()

			if lang == "" {
				lang = "ru"
			}
			h.sendMessage(chatID, translations[lang]["err_files"])
			return
		}
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callbackData = update.CallbackQuery.Data

		// Ответ на CallbackQuery для снятия "часиков" с кнопки
		_, err := h.bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
		if err != nil {
			log.Printf("[WARNING] Failed to acknowledge callback: %v", err)
		}
	} else {
		return
	}

	// Извлекаем состояние как указатель
	state := h.getOrCreateState(chatID)

	// Блокируем сессию на время обработки текущего апдейта
	state.mu.Lock()
	defer state.mu.Unlock()

	// Сброс состояния
	if text == "/start" {
		state.Step = StateLanguage
		state.Language = ""
		state.Answers = make(map[string]string)
		h.sendLanguageSelection(chatID)
		return
	}

	// Маршрутизатор (FSM)
	switch state.Step {
	case StateLanguage:
		if callbackData == "lang_ru" || callbackData == "lang_kk" {
			state.Language = callbackData[5:]
			state.Step = StateMenu
			h.sendMenu(chatID, state.Language)
		} else {
			// Если пользователь пишет текст, когда нужно выбрать язык
			lang := "ru" // fallback
			h.sendMessage(chatID, translations[lang]["err_lang"]+"\n"+translations["kk"]["err_lang"])
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
			state.Step = StateAuditQuestions
			state.QuestionIndex = 0
			h.sendAuditQuestion(chatID, state)
		} else {
			h.sendMessage(chatID, translations[state.Language]["err_bin"])
		}

	case StateAuditQuestions:
		// Принимаем только системные ключи (ans_yes, ans_no, ans_idk)
		if callbackData == "ans_yes" || callbackData == "ans_no" || callbackData == "ans_idk" {
			// В БД сохраняем читаемый вариант на основе ключа, либо сам ключ (оставил системные ключи для единообразия в БД)
			qKey := fmt.Sprintf("q%d", state.QuestionIndex+1)
			state.Answers[qKey] = callbackData
			state.QuestionIndex++

			if state.QuestionIndex < len(auditQuestions[state.Language]) {
				h.sendAuditQuestion(chatID, state)
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
		if err != nil {
			log.Printf("[ERROR] SaveAuditRecord: %v", err)
			h.sendMessage(chatID, translations[state.Language]["err_save"])
		} else {
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

			err := h.repo.SaveAppointment(ctx, chatID, state.TargetManager, state.FullName, state.PhoneNumber, state.Question)
			if err != nil {
				log.Printf("[ERROR] SaveAppointment: %v", err)
				h.sendMessage(chatID, translations[state.Language]["err_save"])
			} else {
				adminChatID := int64(601610)
				adminMsg := fmt.Sprintf("🔔 *Новая запись на прием!*\n\n*К кому:* %s\n*ФИО:* %s\n*Вопрос:* %s\n*Телефон:* %s",
					state.TargetManager, state.FullName, state.Question, state.PhoneNumber)

				msg := tgbotapi.NewMessage(adminChatID, adminMsg)
				msg.ParseMode = "Markdown"

				if _, err := h.bot.Send(msg); err != nil {
					log.Printf("[ERROR] Не удалось отправить алерт: %v", err)
				}

				h.sendMessage(chatID, translations[state.Language]["app_success"])
			}
			state.Step = StateMenu
			h.sendMenu(chatID, state.Language)
		}
	}
	// h.sessions.Store больше не нужен, так как мы работаем с указателем state по ссылке
}

// --- Вспомогательные методы ---

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
	msg := tgbotapi.NewMessage(chatID, text)
	h.bot.Send(msg)
}

func (h *BotHandler) sendDocument(chatID int64, fileURL string) {
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileURL(fileURL))
	if _, err := h.bot.Send(doc); err != nil {
		log.Printf("[ERROR] Failed to send doc to %d: %v", chatID, err)
	}
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

func (h *BotHandler) sendAuditQuestion(chatID int64, state *UserState) {
	lang := state.Language
	index := state.QuestionIndex

	msg := tgbotapi.NewMessage(chatID, auditQuestions[lang][index])
	
	// Используем унифицированные ключи `ans_yes`, `ans_no`, `ans_idk`
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(translations[lang]["btn_yes"], "yes"),
			tgbotapi.NewInlineKeyboardButtonData(translations[lang]["btn_no"], "no"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(translations[lang]["btn_idk"], "idk"),
		),
	)
	h.bot.Send(msg)
}

func (h *BotHandler) isValidBIN(bin string) bool {
	match, _ := regexp.MatchString(`^\d{12}$`, bin)
	return match
}