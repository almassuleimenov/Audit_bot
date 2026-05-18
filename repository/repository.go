package repository
//D:\Project\backend_projects\audit_bot\repository\repository.go
import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BotRepository interface {
	SaveAuditRecord(ctx context.Context, chatID int64, bin string, position string, answers map[string]string, score int) error
	SaveAppointment(ctx context.Context, chatID int64, target string, fullName string, phone string, question string) error
	GetAuditRecords(ctx context.Context, limit int, offset int) ([]AuditRecord, error)
	GetAppointments(ctx context.Context, limit int, offset int) ([]Appointment, error)

	// Методы для динамических вопросов
	GetActiveQuestions(ctx context.Context) ([]SurveyQuestion, error)
	GetAllQuestions(ctx context.Context) ([]SurveyQuestion, error)
	SaveQuestion(ctx context.Context, q *SurveyQuestion) error
	SeedDefaultQuestions(ctx context.Context) error
}

type botRepositoryImpl struct {
	db *gorm.DB
}

func NewBotRepository(db *gorm.DB) BotRepository {
	return &botRepositoryImpl{db: db}
}

func (r *botRepositoryImpl) SaveAuditRecord(ctx context.Context, chatID int64, bin string, position string, answers map[string]string, score int) error {
	answersBytes, err := json.Marshal(answers)
	if err != nil {
		return fmt.Errorf("failed to marshal answers: %w", err)
	}
	record := &AuditRecord{
		TelegramID: chatID,
		BIN:        bin,
		Position:   position,
		Answers:    datatypes.JSON(answersBytes),
		Score:      score,
	}
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *botRepositoryImpl) SaveAppointment(ctx context.Context, chatID int64, target string, fullName string, phone string, question string) error {
	appointment := &Appointment{
		TelegramID:    chatID,
		TargetManager: target,
		FullName:      fullName,
		PhoneNumber:   phone,
		Question:      question,
	}
	return r.db.WithContext(ctx).Create(appointment).Error
}

func (r *botRepositoryImpl) GetAuditRecords(ctx context.Context, limit int, offset int) ([]AuditRecord, error) {
	var records []AuditRecord
	err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Offset(offset).Find(&records).Error
	return records, err
}

func (r *botRepositoryImpl) GetAppointments(ctx context.Context, limit int, offset int) ([]Appointment, error) {
	var appointments []Appointment
	err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Offset(offset).Find(&appointments).Error
	return appointments, err
}

func (r *botRepositoryImpl) GetActiveQuestions(ctx context.Context) ([]SurveyQuestion, error) {
	var questions []SurveyQuestion
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("order_num asc").Find(&questions).Error
	return questions, err
}

func (r *botRepositoryImpl) GetAllQuestions(ctx context.Context) ([]SurveyQuestion, error) {
	var questions []SurveyQuestion
	err := r.db.WithContext(ctx).Order("order_num asc").Find(&questions).Error
	return questions, err
}

func (r *botRepositoryImpl) SaveQuestion(ctx context.Context, q *SurveyQuestion) error {
	return r.db.WithContext(ctx).Save(q).Error
}

// SeedDefaultQuestions заполняет базу стандартными вопросами, если она пуста
func (r *botRepositoryImpl) SeedDefaultQuestions(ctx context.Context) error {
	var count int64
	r.db.Model(&SurveyQuestion{}).Count(&count)
	
	if count > 0 {
		return nil // Вопросы уже есть, ничего не делаем
	}

	log.Println("[INFO] База вопросов пуста. Загружаем вопросы по умолчанию...")

	defaultOptionsRU, _ := json.Marshal([]string{"Да", "Нет", "Затрудняюсь"})
	defaultOptionsKK, _ := json.Marshal([]string{"Иә", "Жоқ", "Қиналамын"})

	questionsRU := []string{
		"1. Были ли сотрудники аудита вежливы, корректны и уважительны в общении в процессе проверки?",
		"2. Возникали ли в процессе проверки ситуации, которые могли носить признаки давления?",
		"3. Предлагал ли кто-либо из сотрудников неофициальное решение по результатам проверки?",
		"4. Были ли попытки получения личной выгоды со стороны проверяющих?",
		"5. Демонстрировали ли аудиторы прозрачность и объективность?",
		"6. Вмешивались ли аудиторы в деятельность организации за рамками компетенции?",
		"7. Придерживались ли сотрудники аудита принципов конфиденциальности?",
		"8. Возникали ли у вас ощущения предвзятого отношения?",
		"9. Оказывалось ли к вам давление с целью скрыть информацию?",
	}

	questionsKK := []string{
		"1. Аудит қызметкерлері тексеру барысында сыпайы, әдепті сөйлесті ме?",
		"2. Тексеру барысында аудиторлар тарапынан қысым көрсету белгілері байқалды ма?",
		"3. Аудит қызметкерлері бейресми шешім ұсынды ма?",
		"4. Тексерушілер тарапынан жеке пайда алу әрекеттері болды ма?",
		"5. Аудиторлар объективтілікті көрсетті ме?",
		"6. Аудиторлар өз құзыретінен тыс ұйымның қызметіне араласты ма?",
		"7. Аудит қызметкерлері құпиялылық қағидаттарын сақтады ма?",
		"8. Аудиторлар сіздің ұйымыңызға біржақты қарайды деген сезім болды ма?",
		"9. Аудит барысында ақпаратты жасыру мақсатында сізге қысым көрсетілді ме?",
	}

	for i := 0; i < len(questionsRU); i++ {
		q := &SurveyQuestion{
			TextRU:    questionsRU[i],
			TextKK:    questionsKK[i],
			OptionsRU: datatypes.JSON(defaultOptionsRU),
			OptionsKK: datatypes.JSON(defaultOptionsKK),
			OrderNum:  i + 1,
			IsActive:  true,
		}
		if err := r.db.Create(q).Error; err != nil {
			return err
		}
	}
	return nil
}