package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BotRepository определяет контракт для слоя работы с базой данных бота.
type BotRepository interface {
	SaveAuditRecord(ctx context.Context, chatID int64, bin string, position string, answers map[string]string, score int) error
	SaveAppointment(ctx context.Context, chatID int64, target string, fullName string, phone string, question string) error
	GetAuditRecords(ctx context.Context, limit int, offset int) ([]AuditRecord, error)
	GetAppointments(ctx context.Context, limit int, offset int) ([]Appointment, error)
}

// botRepositoryImpl реализует интерфейс BotRepository.
type botRepositoryImpl struct {
	db *gorm.DB
}

// NewBotRepository создает новый экземпляр репозитория.
func NewBotRepository(db *gorm.DB) BotRepository {
	return &botRepositoryImpl{
		db: db,
	}
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
		log.Printf("[ERROR] Failed to save audit record for TelegramID %d: %v\n", chatID, err)
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
		log.Printf("[ERROR] Failed to save appointment for TelegramID %d: %v\n", chatID, err)
		return fmt.Errorf("repository.SaveAppointment - db insert failed: %w", err)
	}

	return nil
}

// GetAuditRecords возвращает список аудитов с учетом пагинации (избегаем Out of Memory)
func (r *botRepositoryImpl) GetAuditRecords(ctx context.Context, limit int, offset int) ([]AuditRecord, error) {
	var records []AuditRecord
	if err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch audits: %w", err)
	}
	return records, nil
}

// GetAppointments возвращает список заявок с учетом пагинации
func (r *botRepositoryImpl) GetAppointments(ctx context.Context, limit int, offset int) ([]Appointment, error) {
	var appointments []Appointment
	if err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Offset(offset).Find(&appointments).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch appointments: %w", err)
	}
	return appointments, nil
}
