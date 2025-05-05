package votes_repositories

import (
	"VoteGolang/internals/data/candidate_data"
	"gorm.io/gorm"
)

// VoteRepository manages voting data for candidates.
type VoteRepository interface {
	HasVoted(userID uint, voteType string) (bool, error)
	SaveVote(candidateID uint, userID uint, voteType string) error
}

type voteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) VoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) HasVoted(userID uint, voteType string) (bool, error) {
	var count int64
	err := r.db.Model(&candidate_data.Vote{}).
		Where("user_id = ? AND candidate_type = ?", userID, voteType).
		Count(&count).Error
	return count > 0, err
}

func (r *voteRepository) SaveVote(candidateID uint, userID uint, candidateType string) error {
	vote := &candidate_data.Vote{
		CandidateID:   candidateID,
		UserID:        userID,
		CandidateType: candidateType,
	}
	return r.db.Create(vote).Error
}
