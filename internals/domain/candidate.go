package domain

import (
	"time"

	"gorm.io/gorm"
)

// Candidate represents a candidate for election.
type Candidate struct {
	ID             uint           `gorm:"primaryKey;autoIncrement"`
	Name           string         `gorm:"type:varchar(255);not null"`
	Photo          *string        `gorm:"type:varchar(255)"`
	Education      *string        `gorm:"type:varchar(255)"`
	Age            int            `gorm:"not null"`
	Party          *string        `gorm:"type:varchar(255)"`
	Region         *string        `gorm:"type:varchar(255)"`
	Votes          int            `gorm:"default:0"`
	Type           CandidateType  `gorm:"type:varchar(255);not null"`
	VotingStart    time.Time      `gorm:"type:datetime" example:"2025-05-10T23:59:00+05:00"`
	VotingDeadline time.Time      `gorm:"type:datetime" example:"2025-05-10T23:59:00+05:00"`
	DeletedAt      gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" swaggerignore:"true"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" swaggerignore:"true"`
}

// CandidateRepository retrieves candidate data.
type CandidateRepository interface {
	GetAllByType(candidateType string) ([]Candidate, error)
	GetByID(id uint) (*Candidate, error)
	IncrementVote(id uint) error
	GetAllByTypePaginated(candidateType string, limit, offset int) ([]Candidate, error)
}

type CandidateType string

const (
	Presidential CandidateType = "presidential"
	Deputy       CandidateType = "deputy"
	Manager      CandidateType = "manager"
)

func IsValidCandidateType(t string) bool {
	switch CandidateType(t) {
	case Presidential, Deputy, Manager:
		return true
	default:
		return false
	}
}
