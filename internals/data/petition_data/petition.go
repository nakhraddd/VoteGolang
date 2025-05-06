package petition_data

import (
	"gorm.io/gorm"
	"time"
)

// Petition represents a created petition_data.
type Petition struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" swaggerignore:"true"`
	UserID         uint           `gorm:"type:varchar(255);not null" swaggerignore:"true"`
	Title          string         `gorm:"type:varchar(255);not null"`
	Photo          *string        `gorm:"type:varchar(255)"`
	Description    *string        `gorm:"type:text"`
	VotesInFavor   int            `gorm:"default:0" swaggerignore:"true"`
	VotesAgainst   int            `gorm:"default:0" swaggerignore:"true"`
	Goal           int            `gorm:"not null"`
	VotingDeadline time.Time      `gorm:"type:datetime" example:"2025-05-10T23:59:00+05:00"`
	DeletedAt      gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" swaggerignore:"true"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" swaggerignore:"true"`
}
