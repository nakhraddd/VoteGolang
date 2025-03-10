package models

import "gorm.io/gorm"

type Vote struct {
	gorm.Model
	UserID        uint   `gorm:"index"`
	CandidateName string `gorm:"index"`
}

type PetitionVote struct {
	gorm.Model
	UserID        uint   `gorm:"index"`
	PetitionTitle string `gorm:"index"`
	VoteType      string
}
