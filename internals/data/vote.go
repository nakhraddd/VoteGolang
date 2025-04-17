package data

import (
	"gorm.io/gorm"
	"time"
)

type Vote struct {
	ID            uint   `gorm:"primaryKey;autoIncrement"`
	UserID        uint   `gorm:"not null"`
	CandidateID   uint   `gorm:"not null"`
	CandidateType string `gorm:"type:varchar(50)"`
	DeletedAt     gorm.DeletedAt
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}
