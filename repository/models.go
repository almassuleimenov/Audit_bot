package repository
//D:\Project\backend_projects\audit_bot\repository\models.go
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
	Answers    datatypes.JSON `gorm:"type:jsonb;not null;comment:JSON с ответами на вопросы"`
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

// SurveyQuestion представляет динамический вопрос анкеты с вариантами ответов
type SurveyQuestion struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TextRU    string         `gorm:"type:text;not null;comment:Вопрос на русском" json:"text_ru"`
	TextKK    string         `gorm:"type:text;not null;comment:Вопрос на казахском" json:"text_kk"`
	OptionsRU datatypes.JSON `gorm:"type:jsonb;not null;comment:Варианты на русском (массив строк)" json:"options_ru"`
	OptionsKK datatypes.JSON `gorm:"type:jsonb;not null;comment:Варианты на казахском (массив строк)" json:"options_kk"`
	OrderNum  int            `gorm:"not null;comment:Порядковый номер" json:"order_num"`
	IsActive  bool           `gorm:"default:true;comment:Активен ли вопрос" json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
}