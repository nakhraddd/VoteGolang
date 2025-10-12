package domain

import (
	"time"

	"gorm.io/gorm"
)

// Petition represents a created petition.
type Petition struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" swaggerignore:"true"`
	UserID         uint           `gorm:"not null" swaggerignore:"true"`
	Title          string         `gorm:"type:varchar(255);not null"`
	Photo          *string        `gorm:"type:varchar(255)"`
	Description    *string        `gorm:"type:text"`
	VotesInFavor   int            `gorm:"default:0" swaggerignore:"true"`
	VotesAgainst   int            `gorm:"default:0" swaggerignore:"true"`
	Goal           int            `gorm:"not null"`
	VotingDeadline time.Time      `json:"voting_deadline" gorm:"type:datetime" example:"2025-05-10T23:59:00+05:00"`
	DeletedAt      gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" swaggerignore:"true"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" swaggerignore:"true"`
}

type PetitionRepository interface {
	Create(petition *Petition) error
	GetAll() ([]Petition, error)
	GetAllPaginated(limit, offset int) ([]Petition, error)
	GetByID(id uint) (*Petition, error)
	VoteInFavor(id uint) error
	VoteAgainst(id uint) error
	Delete(id uint) error
}

type PetitionVote struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	UserID     uint `gorm:"not null" swaggerignore:"true"`
	PetitionID uint
	VoteType   VoteType       `gorm:"type:varchar(255);not null"`
	DeletedAt  gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
}

// PetitionVoteRequest represents the body to petition for a petition petition request.
type PetitionVoteRequest struct {
	UserId     uint `json:"user_id" swaggerignore:"true"`
	PetitionID uint `json:"petition_id"`
	// Enum values: favor, against
	VoteType VoteType `json:"vote_type" enum:"favor,against" example:"favor, against"`
}

type PetitionIDRequest struct {
	ID uint `json:"id"`
}

type VoteType string

const (
	Favor   VoteType = "favor"
	Against VoteType = "against"
)

func IsValidVoteType(v string) bool {
	return v == string(Favor) || v == string(Against)
}
