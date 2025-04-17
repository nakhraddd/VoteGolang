package data

import (
	"gorm.io/gorm"
	"time"
)

type GeneralNews struct {
	ID        uint    `gorm:"primaryKey;autoIncrement"`
	Title     string  `gorm:"type:varchar(255);not null"`
	Paragraph *string `gorm:"type:text"`
	Photo     *string `gorm:"type:varchar(255)"`
	DeletedAt gorm.DeletedAt
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
