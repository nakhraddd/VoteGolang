package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// User represents a registered user in the system.
type User struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" swaggerignore:"true"`
	Username      string         `gorm:"type:varchar(100);not null;unique" example:"beks"`
	Email         string         `gorm:"type:varchar(100);not null;unique" example:"zhaslanbeksultan@gmail.com"`
	EmailVerified bool           `gorm:"default:false" swaggerignore:"true"`
	UserFullName  *string        `gorm:"type:varchar(110)" example:"Beksultan Zhaslan"`
	Password      string         `gorm:"type:varchar(255);not null" example:"$Password123"`
	BirthDate     *time.Time     `example:"2004-07-16T00:00:00Z"`
	Address       *string        `gorm:"type:text" example:"59, Tole Bi St, Almaty"`
	DeletedAt     gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" swaggerignore:"true"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" swaggerignore:"true"`
	RoleID        uint           `gorm:"not null" example:"2"`
	Role          Role           `gorm:"foreignKey:RoleID" swaggerignore:"true"`
}

// UserRepository handles database operations related to users.
type UserRepository interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
	MarkEmailVerified(ctx context.Context, userID uint) error
	GetByEmail(email string) (*User, error)
	DeleteUnverifiedUser(cutoff time.Time) (int64, error)
}
