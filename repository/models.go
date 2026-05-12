package repository

import (
	"time"

	"gorm.io/datatypes"
)

// AuditRecord представляет таблицу для хранения результатов анкетирования.
type AuditRecord struct {
	ID         uint           `gorm:"primaryKey"`
	TelegramID int64          `gorm:"index;not null;comment:ID пользователя в Telegram"`
	BIN        string         `gorm:"type:varchar(12);not null;comment:БИН организации"`
	Position   string         `gorm:"type:varchar(255);not null;comment:Должность сотрудника"`
	Answers    datatypes.JSON `gorm:"type:jsonb;not null;comment:JSON с ответами на 9 вопросов"`
	Score      int            `gorm:"not null;comment:Оценка от 1 до 5"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;index"`
}

// Appointment представляет таблицу для записи на онлайн-прием.
type Appointment struct {
	ID            uint      `gorm:"primaryKey"`
	TelegramID    int64     `gorm:"index;not null"`
	TargetManager string    `gorm:"type:varchar(255);not null;comment:Выбранный руководитель"`
	FullName      string    `gorm:"type:varchar(255);not null;comment:ФИО заявителя"`
	PhoneNumber   string    `gorm:"type:varchar(20);not null;comment:Контактный номер"`
	Question      string    `gorm:"type:text;not null;comment:Суть вопроса"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index"`
}