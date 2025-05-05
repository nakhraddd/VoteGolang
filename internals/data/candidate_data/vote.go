package candidate_data

import (
	"gorm.io/gorm"
	"time"
)

// Vote represents a petition_data on a candidate.
type Vote struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" swaggerignore:"true"`
	UserID        uint           `gorm:"not null" swaggerignore:"true"`
	CandidateID   uint           `gorm:"not null"`
	CandidateType string         `gorm:"type:varchar(50)"`
	DeletedAt     gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" swaggerignore:"true"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" swaggerignore:"true"`
}
