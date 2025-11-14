package repositories

import (
	"VoteGolang/internals/domain"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	// Use OnConflict to make it idempotent
	// This will do nothing if the record already exists
	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "candidate_type"}},
		DoNothing: true,
	}).Create(vote)

	if result.Error != nil {
		return result.Error
	}

	// If RowsAffected is 0, the vote already existed
	// This is not an error for idempotency
	return nil
}

func (r *voteGormRepository) VoteWithTransaction(candidateID uint, userID uint, candidateType string, afterSave func() error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Check if already voted (with row lock to prevent race conditions)
		var existingVote domain.Vote
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND candidate_type = ?", userID, candidateType).
			First(&existingVote).Error

		if err == nil {
			// Vote already exists - idempotent response
			return errors.New("already voted for this category")
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Real database error
			return err
		}

		// Create the vote record
		vote := &domain.Vote{
			CandidateID:   candidateID,
			UserID:        userID,
			CandidateType: domain.CandidateType(candidateType),
		}

		if err := tx.Create(vote).Error; err != nil {
			return err
		}

		// Execute callback (increment vote count, blockchain, etc.)
		if afterSave != nil {
			if err := afterSave(); err != nil {
				return err
			}
		}

		return nil
	})
}
