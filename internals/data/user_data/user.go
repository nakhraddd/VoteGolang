package user_data

import (
	"time"

	"gorm.io/gorm"
)

// User represents a registered user in the system.
type User struct {
	ID           uint    `gorm:"primaryKey;autoIncrement"`
	Username     string  `gorm:"type:varchar(100);not null;unique"`
	UserFullName *string `gorm:"type:varchar(110)"`
	Password     string  `gorm:"type:varchar(255);not null"`
	BirthDate    *time.Time
	Address      *string        `gorm:"type:text"`
	DeletedAt    gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" swaggerignore:"true"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
}
