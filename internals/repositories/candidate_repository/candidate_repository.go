package candidate_repository

import (
	"VoteGolang/internals/data/candidate_data"
	"gorm.io/gorm"
)

// CandidateRepository retrieves candidate data.
type CandidateRepository interface {
	GetAllByType(candidateType string) ([]candidate_data.Candidate, error)
	GetByID(id uint) (*candidate_data.Candidate, error)
	IncrementVote(id uint) error
	GetAllByTypePaginated(candidateType string, limit, offset int) ([]candidate_data.Candidate, error)
}

type candidateRepository struct {
	db *gorm.DB
}

func NewCandidateRepository(db *gorm.DB) CandidateRepository {
	return &candidateRepository{db: db}
}

func (r *candidateRepository) GetAllByTypePaginated(candidateType string, limit, offset int) ([]candidate_data.Candidate, error) {
	var candidates []candidate_data.Candidate
	err := r.db.
		Where("type = ?", candidateType).
		Limit(limit).
		Offset(offset).
		Find(&candidates).Error
	return candidates, err
}

func (r *candidateRepository) GetAllByType(candidateType string) ([]candidate_data.Candidate, error) {
	var candidates []candidate_data.Candidate
	err := r.db.Where("type = ?", candidateType).Find(&candidates).Error
	return candidates, err
}

func (r *candidateRepository) GetByID(id uint) (*candidate_data.Candidate, error) {
	var candidate candidate_data.Candidate
	err := r.db.First(&candidate, id).Error
	if err != nil {
		return nil, err
	}
	return &candidate, nil
}

func (r *candidateRepository) IncrementVote(id uint) error {
	return r.db.Model(&candidate_data.Candidate{}).
		Where("id = ?", id).
		UpdateColumn("votes", gorm.Expr("votes + ?", 1)).Error
}
