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

// SaveAuditRecord конвертирует мапу ответов в JSONB и сохраняет запись в БД.
func (r *botRepositoryImpl) SaveAuditRecord(ctx context.Context, chatID int64, bin string, position string, answers map[string]string, score int) error {
	// Сериализация мапы в байты для типа datatypes.JSON (O(N) time complexity)
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

	// Используем WithContext для контроля таймаутов
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		log.Printf("[ERROR] Failed to save audit record for TelegramID %d: %v\n", chatID, err)
		return fmt.Errorf("repository.SaveAuditRecord - db insert failed: %w", err)
	}

	log.Printf("[INFO] Successfully saved audit record for TelegramID %d\n", chatID)
	return nil
}

// SaveAppointment атомарно сохраняет заявку на онлайн-прием.
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

	log.Printf("[INFO] Successfully saved appointment for TelegramID %d\n", chatID)
	return nil
}
