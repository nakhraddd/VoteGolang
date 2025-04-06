package data

import "gorm.io/gorm"

type GeneralNews struct {
	gorm.Model
	Title     string `json:"title"`
	Paragraph string `json:"paragraph"`
	Photo     string `json:"photo"`
}
