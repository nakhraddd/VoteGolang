package repositories

import (
	petition_data2 "VoteGolang/internals/domain"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type petitionVoteGormRepository struct {
	db *gorm.DB
}

func NewPetitionVoteRepository(db *gorm.DB) petition_data2.PetitionVoteRepository {
	return &petitionVoteGormRepository{db: db}
}

func (r *petitionVoteGormRepository) CreateVote(vote *petition_data2.PetitionVote) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Use OnConflict for idempotency
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "petition_id"}},
			DoNothing: true,
		}).Create(vote)

		if result.Error != nil {
			return result.Error
		}

		// If no rows affected, vote already existed (idempotent)
		if result.RowsAffected == 0 {
			return nil
		}

		var column string
		switch vote.VoteType {
		case petition_data2.Favor:
			column = "votes_in_favor"
		case petition_data2.Against:
			column = "votes_against"
		default:
			return nil
		}

		return tx.Model(&petition_data2.Petition{}).
			Where("id = ?", vote.PetitionID).
			UpdateColumn(column, gorm.Expr(column+" + ?", 1)).
			Error
	})
}

func (r *petitionVoteGormRepository) HasUserVoted(userID uint, petitionID uint) (bool, error) {
	var count int64
	err := r.db.Model(&petition_data2.PetitionVote{}).
		Where("user_id = ? AND petition_id = ?", userID, petitionID).
		Count(&count).Error

	return count > 0, err
}

// VoteWithTransaction ensures atomicity and idempotency with row locking
func (r *petitionVoteGormRepository) VoteWithTransaction(userID uint, petitionID uint, voteType petition_data2.VoteType, afterSave func() error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Check if already voted with row lock to prevent race conditions
		var existingVote petition_data2.PetitionVote
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND petition_id = ?", userID, petitionID).
			First(&existingVote).Error

		if err == nil {
			// Vote already exists - return idempotent error
			return errors.New("user has already voted")
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Real database error
			return err
		}

		// Create the vote record
		vote := &petition_data2.PetitionVote{
			UserID:     userID,
			PetitionID: petitionID,
			VoteType:   voteType,
		}

		if err := tx.Create(vote).Error; err != nil {
			return err
		}

		// Execute callback (update vote counts, blockchain, etc.)
		if afterSave != nil {
			if err := afterSave(); err != nil {
				return err
			}
		}

		return nil
	})
}
