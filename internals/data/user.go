package data

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID           uint    `gorm:"primaryKey;autoIncrement"`
	Username     string  `gorm:"type:varchar(100);not null;unique"`
	UserFullName *string `gorm:"type:varchar(100)"`
	Password     string  `gorm:"type:varchar(255);not null"`
	BirthDate    *time.Time
	Address      *string `gorm:"type:text"`
	DeletedAt    gorm.DeletedAt
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
