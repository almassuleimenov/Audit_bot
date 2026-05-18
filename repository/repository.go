package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BotRepository interface {
	SaveAuditRecord(ctx context.Context, chatID int64, bin string, position string, answers map[string]string, score int) error
	SaveAppointment(ctx context.Context, chatID int64, target string, fullName string, phone string, question string) error
	GetAuditRecords(ctx context.Context, limit int, offset int) ([]AuditRecord, error)
	GetAppointments(ctx context.Context, limit int, offset int) ([]Appointment, error)

	// Новые методы для вопросов
	GetActiveQuestions(ctx context.Context) ([]SurveyQuestion, error)
	SaveQuestion(ctx context.Context, q *SurveyQuestion) error
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
		return fmt.Errorf("repository.SaveAuditRecord - failed to marshal answers: %w", err)
	}
	record := &AuditRecord{
		TelegramID: chatID,
		BIN:        bin,
		Position:   position,
		Answers:    datatypes.JSON(answersBytes),
		Score:      score,
	}
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("repository.SaveAuditRecord - db insert failed: %w", err)
	}
	return nil
}

func (r *botRepositoryImpl) SaveAppointment(ctx context.Context, chatID int64, target string, fullName string, phone string, question string) error {
	appointment := &Appointment{
		TelegramID:    chatID,
		TargetManager: target,
		FullName:      fullName,
		PhoneNumber:   phone,
		Question:      question,
	}
	if err := r.db.WithContext(ctx).Create(appointment).Error; err != nil {
		return fmt.Errorf("repository.SaveAppointment - db insert failed: %w", err)
	}
	return nil
}

func (r *botRepositoryImpl) GetAuditRecords(ctx context.Context, limit int, offset int) ([]AuditRecord, error) {
	var records []AuditRecord
	if err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch audits: %w", err)
	}
	return records, nil
}

func (r *botRepositoryImpl) GetAppointments(ctx context.Context, limit int, offset int) ([]Appointment, error) {
	var appointments []Appointment
	if err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Offset(offset).Find(&appointments).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch appointments: %w", err)
	}
	return appointments, nil
}

func (r *botRepositoryImpl) GetActiveQuestions(ctx context.Context) ([]SurveyQuestion, error) {
	var questions []SurveyQuestion
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("order_num asc").Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active questions: %w", err)
	}
	return questions, nil
}

func (r *botRepositoryImpl) SaveQuestion(ctx context.Context, q *SurveyQuestion) error {
	if err := r.db.WithContext(ctx).Save(q).Error; err != nil {
		return fmt.Errorf("failed to save question: %w", err)
	}
	return nil
}
