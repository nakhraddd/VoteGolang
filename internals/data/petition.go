package data

import (
	"gorm.io/gorm"
	"time"
)

// Petition represents a created petition.
type Petition struct {
	ID           uint    `gorm:"primaryKey;autoIncrement"`
	UserID       uint    `gorm:"type:varchar(255);not null"`
	Title        string  `gorm:"type:varchar(255);not null"`
	Photo        *string `gorm:"type:varchar(255)"`
	Description  *string `gorm:"type:text"`
	VotesInFavor int     `gorm:"default:0"`
	VotesAgainst int     `gorm:"default:0"`
	DeletedAt    gorm.DeletedAt
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
