package repositories

import (
	"VoteGolang/internals/domain"

	"gorm.io/gorm"
)

type voteGormRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) domain.VoteRepository {
	return &voteGormRepository{db: db}
}

func (r *voteGormRepository) HasVoted(userID uint, voteType string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Vote{}).
		Where("user_id = ? AND candidate_type = ?", userID, voteType).
		Count(&count).Error
	return count > 0, err
}

func (r *voteGormRepository) SaveVote(candidateID uint, userID uint, candidateType string) error {
	vote := &domain.Vote{
		CandidateID:   candidateID,
		UserID:        userID,
		CandidateType: domain.CandidateType(candidateType),
	}
	return r.db.Create(vote).Error
}
