package data

import (
	"gorm.io/gorm"
	"time"
)

// Candidate represents a candidate for election.
type Candidate struct {
	ID        uint    `gorm:"primaryKey;autoIncrement"`
	Name      string  `gorm:"type:varchar(255);not null"`
	Photo     *string `gorm:"type:varchar(255)"`
	Education *string `gorm:"type:varchar(255)"`
	Age       int     `gorm:"not null"`
	Party     *string `gorm:"type:varchar(255)"`
	Region    *string `gorm:"type:varchar(255)"`
	Votes     int     `gorm:"default:0"`
	Type      string  `gorm:"type:varchar(255);not null"`
	DeletedAt gorm.DeletedAt
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
