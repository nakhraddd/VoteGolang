package data

import (
	"gorm.io/gorm"
	"time"
)

// PetitionVote represents a vote on a petition.
type PetitionVote struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	UserID     uint `gorm:"type:varchar(255);not null"`
	PetitionID uint
	VoteType   string         `gorm:"type:varchar(255);not null"`
	DeletedAt  gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
}
