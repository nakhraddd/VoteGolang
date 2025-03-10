package models

import "gorm.io/gorm"

type Petition struct {
	gorm.Model
	Title        string `json:"title"`
	Photo        string `json:"photo"`
	Description  string `json:"description"`
	VotesInFavor int    `json:"votes"`
	VotesAgainst int    `json:"votes_against"`
}
