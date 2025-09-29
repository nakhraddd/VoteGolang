package domain

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

// UserRepository handles database operations related to users.
type UserRepository interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
}
