package repositories

import (
	"VoteGolang/internals/data/models"
	"gorm.io/gorm"
)

type CandidateRepository interface {
	GetAllByType(candidateType string) ([]models.Candidate, error)
	GetByID(id uint) (*models.Candidate, error)
	IncrementVote(id uint) error
}

type candidateRepository struct {
	db *gorm.DB
}

func (c *candidateRepository) GetAllByType(candidateType string) ([]models.Candidate, error) {
	var candidates []models.Candidate
	err := c.db.Where("type = ?", candidateType).Find(&candidates).Error
	return candidates, err
}

func (c *candidateRepository) GetByID(id uint) (*models.Candidate, error) {
	var candidate models.Candidate
	err := c.db.First(&candidate, id).Error
	if err != nil {
		return nil, err
	}
	return &candidate, nil
}

func (c *candidateRepository) IncrementVote(id uint) error {
	return c.db.Model(&models.Candidate{}).
		Where("id = ?", id).
		UpdateColumn("votes", gorm.Expr("votes + ?", 1)).Error
}

func NewCandidateRepository(db *gorm.DB) CandidateRepository {
	return &candidateRepository{db: db}
}

func (r *candidateRepository) GetAllByType(candidateType string) ([]models.Candidate, error) {
	var candidates []models.Candidate
	err := r.db.Where("type = ?", candidateType).Find(&candidates).Error
	return candidates, err
}

func (r *candidateRepository) GetByID(id uint) (*models.Candidate, error) {
	var candidate models.Candidate
	err := r.db.First(&candidate, id).Error
	if err != nil {
		return nil, err
	}
	return &candidate, nil
}

func (r *candidateRepository) IncrementVote(id uint) error {
	return r.db.Model(&models.Candidate{}).
		Where("id = ?", id).
		UpdateColumn("votes", gorm.Expr("votes + ?", 1)).Error
}
