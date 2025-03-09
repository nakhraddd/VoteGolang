package models

import "gorm.io/gorm"

type Vote struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	Candidate string `gorm:"not null"`
	Category  string `gorm:"not null"` // "President", "Deputy", etc.
}
