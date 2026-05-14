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
	StateLanguage = iota // НОВЫЙ ШАГ: Выбор языка
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
	Step     int
	Language string // "ru" или "kz"
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

// Словари для интерфейса (О(1) time complexity)
var uiTexts = map[string]map[string]string{
	"ru": {
		"welcome":     "Вас приветствует Чат-бот Департамента внутреннего государственного аудита. Выберите раздел:",
		"menu_audit":  "Анкетирование",
		"menu_app":    "Онлайн-приемная",
		"menu_ethics": "Уполномоченный по этике",
		"menu_press":  "Пресс-центр",
	},
	"kz": {
		"welcome":     "Ішкі мемлекеттік аудит департаментінің чат-ботына қош келдіңіз. Бөлімді таңдаңыз:",
		"menu_audit":  "Сауалнама",
		"menu_app":    "Онлайн-қабылдау",
		"menu_ethics": "Әдеп жөніндегі уәкіл",
		"menu_press":  "Баспасөз орталығы",
	},
}

// Вопросы для аудита (пока на русском, позже можно сделать такой же map для kz)
var auditQuestions = []string{
	"1. Были ли сотрудники аудита вежливы, корректны и уважительны в общении в процессе проверки?",
	"2. Возникали ли в процессе проверки ситуации, которые, по вашему мнению, могли носить признаки давления или нарушения этических норм со стороны аудиторов?",
	"3. Предлагал ли кто-либо из сотрудников аудита неофициальное или \"альтернативное\" решение по вопросам, связанным с результатами проверки?",
	"4. Были ли попытки получения личной выгоды (подарки, деньги, услуги и т. д.) со стороны проверяющих?",
	"5. Демонстрировали ли аудиторы прозрачность и объективность в оценке проверяемой информации?",
	"6. Вмешивались ли аудиторы в деятельность организации за рамками своей компетенции?",
	"7. Придерживались ли сотрудники аудита принципов конфиденциальности при обращении с внутренней документацией и информацией?",
	"8. Возникали ли у вас ощущения, что аудиторы предвзято относятся к вашей организации или отдельным сотрудникам?",
	"9. Оказывалось ли к вам или вашим коллегам давление с целью изменить или скрыть какую-либо информацию в ходе аудита?",
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
	if update.Message != nil && (update.Message.Photo != nil || update.Message.Document != nil || update.Message.Video != nil || update.Message.Audio != nil) {
		h.sendMessage(update.Message.Chat.ID, "Бот не обрабатывает файлы. Пожалуйста, оставьте контактные данные.")
		return
	}

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
			// Сохраняем язык (убираем префикс "lang_")
			state.Language = callbackData[5:]
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
			h.sendMessage(chatID, "Пожалуйста, укажите Ваш БИН (12 цифр):")
		}

	case StateAuditBin:
		if h.isValidBIN(text) {
			state.BIN = text
			state.Step = StateAuditQuestions
			state.QuestionIndex = 0
			h.sendAuditQuestion(chatID, state.QuestionIndex)
		} else {
			h.sendMessage(chatID, "Некорректный БИН. Введите ровно 12 цифр:")
		}

	case StateAuditQuestions:
		if callbackData != "" {
			qKey := fmt.Sprintf("q%d", state.QuestionIndex+1)
			state.Answers[qKey] = callbackData
			state.QuestionIndex++

			if state.QuestionIndex < len(auditQuestions) {
				h.sendAuditQuestion(chatID, state.QuestionIndex)
			} else {
				state.Step = StateAuditScore
				h.sendMessage(chatID, "Благодарим за участие. Оцените аудитора от 1 до 5:")
			}
		}

	case StateAuditScore:
		score, err := strconv.Atoi(text)
		if err != nil || score < 1 || score > 5 {
			h.sendMessage(chatID, "Пожалуйста, введите число от 1 до 5.")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = h.repo.SaveAuditRecord(ctx, chatID, state.BIN, state.Position, state.Answers, score)
		if err != nil {
			log.Printf("[ERROR] SaveAuditRecord: %v", err)
			h.sendMessage(chatID, "Произошла ошибка при сохранении. Попробуйте позже.")
		} else {
			h.sendMessage(chatID, "Принято ✅. Ваши ответы сохранены.")
		}

		state.Step = StateMenu
		h.sendMenu(chatID, state.Language)

	case StateAppManager:
		if callbackData != "" {
			state.TargetManager = callbackData
			state.Step = StateAppFIO
			h.sendMessage(chatID, "Укажите ваше ФИО:")
		}

	case StateAppFIO:
		if text != "" {
			state.FullName = text
			state.Step = StateAppQuestion
			h.sendMessage(chatID, "Характер вопроса кратко:")
		}

	case StateAppQuestion:
		if text != "" {
			state.Question = text
			state.Step = StateAppPhone
			h.sendMessage(chatID, "Введите номер телефона для обратной связи:")
		}

	case StateAppPhone:
		if text != "" {
			state.PhoneNumber = text

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := h.repo.SaveAppointment(ctx, chatID, state.TargetManager, state.FullName, state.PhoneNumber, state.Question)
			if err != nil {
				log.Printf("[ERROR] SaveAppointment: %v", err)
				h.sendMessage(chatID, "Произошла ошибка при записи. Попробуйте позже.")
			} else {
				adminChatID := int64(601610)
				adminMsg := fmt.Sprintf("🔔 *Новая запись на прием!*\n\n*К кому:* %s\n*ФИО:* %s\n*Вопрос:* %s\n*Телефон:* %s",
					state.TargetManager, state.FullName, state.Question, state.PhoneNumber)

				msg := tgbotapi.NewMessage(adminChatID, adminMsg)
				msg.ParseMode = "Markdown"

				if _, err := h.bot.Send(msg); err != nil {
					log.Printf("[ERROR] Не удалось отправить алерт админам: %v", err)
				}

				h.sendMessage(chatID, "Вы успешно записались на прием! С Вами свяжутся в ближайшее время.")
			}
			state.Step = StateMenu
			h.sendMenu(chatID, state.Language)
		}
	}

	h.sessions.Store(chatID, state)
}

// --- Вспомогательные методы ---

func (h *BotHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	h.bot.Send(msg)
}

func (h *BotHandler) sendDocument(chatID int64, fileURL string) {
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileURL(fileURL))
	if _, err := h.bot.Send(doc); err != nil {
		log.Printf("[ERROR] Failed to send document to %d: %v", chatID, err)
	}
}

// НОВЫЙ МЕТОД: Генерация меню выбора языка
func (h *BotHandler) sendLanguageSelection(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите язык / Тілді таңдаңыз:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_ru"),
			tgbotapi.NewInlineKeyboardButtonData("🇰🇿 Қазақша", "lang_kz"),
		),
	)
	h.bot.Send(msg)
}

// Генерация главного меню (теперь принимает язык)
func (h *BotHandler) sendMenu(chatID int64, lang string) {
	// Если язык не задан, ставим дефолтный
	if lang == "" {
		lang = "ru"
	}

	msg := tgbotapi.NewMessage(chatID, uiTexts[lang]["welcome"])

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(uiTexts[lang]["menu_audit"], "menu_audit")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(uiTexts[lang]["menu_app"], "menu_app")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(uiTexts[lang]["menu_ethics"], "menu_ethics")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(uiTexts[lang]["menu_press"], "menu_press")),
	)

	h.bot.Send(msg)
}

func (h *BotHandler) handleMenuChoice(chatID int64, data string, state *UserState) {
	switch data {
	case "menu_press":
		msg := tgbotapi.NewMessage(chatID, "Пресс-центр Комитета:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL("Новости", "https://www.gov.kz/memleket/entities/kvga/press?lang=ru")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL("График", "https://www.gov.kz/memleket/entities/kvga/about/structure/departments/activity/4728/1?lang=ru")),
		)
		h.bot.Send(msg)
		h.sendMenu(chatID, state.Language)

	case "menu_ethics":
		h.sendDocument(chatID, "https://robochat.storage.yandexcloud.net/attachments/day/20284/421499/file/OLnAb2ZL/%D3%98%D0%B4%D0%B5%D0%BF%20%D0%B3%D1%80%D0%B0%D1%84%D0%B8%D0%BA.pdf")
		h.sendMessage(chatID, "Уполномоченный по этике Департамента")
		h.sendMenu(chatID, state.Language)

	case "menu_audit":
		state.Step = StateAuditWarning
		h.sendDocument(chatID, "https://robochat.storage.yandexcloud.net/attachments/day/20285/421499/file/gawYYoV4/%D0%9C%D0%B5%D1%82%D0%BE%D0%B4%D0%B8%D1%87%D0%BA%D0%B0%20%D0%B4%D0%BB%D1%8F%20%D1%81%D0%BE%D1%82%D1%80%20%281%29.pdf")

		// Здесь позже тоже можно вынести текст в `uiTexts` для перевода
		msg := tgbotapi.NewMessage(chatID, "Анкета предназначена для мониторинга соблюдения сотрудниками... Заведомо ложные ответы влекут ответственность (ст. 419, 274 УК РК).")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Ознакомлен ✅", "audit_agree"),
				tgbotapi.NewInlineKeyboardButtonData("Не ознакомлен ❌", "audit_disagree"),
			),
		)
		h.bot.Send(msg)

	case "menu_app":
		state.Step = StateAppManager
		msg := tgbotapi.NewMessage(chatID, "Выберите руководителя для записи:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Қабдыраш Б.С.", "manager_kabdrash")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Мұсабек А.М.", "manager_musabek")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Джумагулов М.Б.", "manager_dzhumagulov")),
		)
		h.bot.Send(msg)
	}
}

func (h *BotHandler) handleAuditWarning(chatID int64, data string, messageID int, state *UserState) {
	if data == "audit_agree" {
		state.Step = StateAuditPosition
		h.sendMessage(chatID, "Пожалуйста, укажите Вашу должность для продолжения:")
	} else if data == "audit_disagree" {
		h.sendMenu(chatID, state.Language)
	}
}

func (h *BotHandler) sendAuditQuestion(chatID int64, index int) {
	msg := tgbotapi.NewMessage(chatID, auditQuestions[index])
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", "Да"),
			tgbotapi.NewInlineKeyboardButtonData("Нет", "Нет"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Затрудняюсь ответить", "Затрудняюсь ответить"),
		),
	)
	h.bot.Send(msg)
}

func (h *BotHandler) isValidBIN(bin string) bool {
	match, _ := regexp.MatchString(`^\d{12}$`, bin)
	return match
}
