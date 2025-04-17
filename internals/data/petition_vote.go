package data

import (
	"gorm.io/gorm"
	"time"
)

type PetitionVote struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	UserID     uint `gorm:"type:varchar(255);not null"`
	PetitionID uint
	VoteType   string `gorm:"type:varchar(255);not null"`
	DeletedAt  gorm.DeletedAt
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}
