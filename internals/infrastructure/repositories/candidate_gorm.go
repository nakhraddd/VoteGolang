package repositories

import (
	"VoteGolang/internals/domain"

	"gorm.io/gorm"
)

type candidateGormRepository struct {
	db *gorm.DB
}

func NewCandidateRepository(db *gorm.DB) domain.CandidateRepository {
	return &candidateGormRepository{db: db}
}

func (r *candidateGormRepository) Create(candidate *domain.Candidate) error {
	return r.db.Create(candidate).Error
}

func (r *candidateGormRepository) GetAllByTypePaginated(candidateType string, limit, offset int) ([]domain.Candidate, error) {
	var candidates []domain.Candidate
	err := r.db.
		Where("type = ?", candidateType).
		Limit(limit).
		Offset(offset).
		Find(&candidates).Error
	return candidates, err
}

func (r *candidateGormRepository) GetAllByType(candidateType string) ([]domain.Candidate, error) {
	var candidates []domain.Candidate
	err := r.db.Where("type = ?", candidateType).Find(&candidates).Error
	return candidates, err
}

func (r *candidateGormRepository) GetByID(id uint) (*domain.Candidate, error) {
	var candidate domain.Candidate
	err := r.db.First(&candidate, id).Error
	if err != nil {
		return nil, err
	}
	return &candidate, nil
}

func (r *candidateGormRepository) IncrementVote(id uint) error {
	return r.db.Model(&domain.Candidate{}).
		Where("id = ?", id).
		UpdateColumn("votes", gorm.Expr("votes + ?", 1)).Error
}
