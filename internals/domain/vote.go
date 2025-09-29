package domain

import (
	"time"

	"gorm.io/gorm"
)

// Vote represents a petition on a candidate.
type Vote struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" swaggerignore:"true"`
	UserID        uint           `gorm:"not null" swaggerignore:"true"`
	CandidateID   uint           `gorm:"not null"`
	CandidateType string         `gorm:"type:varchar(50)"`
	DeletedAt     gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" swaggerignore:"true"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" swaggerignore:"true"`
}

// VoteRequest represents the body to vote_request for a candidate.
// swagger:model
type VoteRequest struct {
	CandidateID   uint   `json:"candidate_id"`
	CandidateType string `json:"candidate_type" enums:"presidential,deputy,manager" example:"presidential, deputy, manager"`
}

// VoteRepository manages voting data for candidates.
type VoteRepository interface {
	HasVoted(userID uint, voteType string) (bool, error)
	SaveVote(candidateID uint, userID uint, voteType string) error
}

// PetitionVoteRepository manages voting data for petitions.
type PetitionVoteRepository interface {
	CreateVote(vote *PetitionVote) error
	HasUserVoted(userID uint, petitionID uint) (bool, error)
}
