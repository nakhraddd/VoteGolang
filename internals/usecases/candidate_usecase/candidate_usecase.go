package candidate_usecase

import (
	"VoteGolang/internals/data/candidate_data"
	"VoteGolang/internals/repositories/candidate_repository"
	"VoteGolang/internals/repositories/votes_repositories"
	"errors"
	"time"
)

// CandidateUseCase handles business logic related to election candidates.
type CandidateUseCase struct {
	CandidateRepo candidate_repository.CandidateRepository
	VoteRepo      votes_repositories.VoteRepository
}

func NewCandidateUseCase(cRepo candidate_repository.CandidateRepository, vRepo votes_repositories.VoteRepository) *CandidateUseCase {
	return &CandidateUseCase{
		CandidateRepo: cRepo,
		VoteRepo:      vRepo,
	}
}
func (uc *CandidateUseCase) GetAllByTypePaginated(candidateType string, limit, offset int) ([]candidate_data.Candidate, error) {
	return uc.CandidateRepo.GetAllByTypePaginated(candidateType, limit, offset)
}

// GetAllByType returns a list of candidates filtered by type.
func (uc *CandidateUseCase) GetAllByType(candidateType string) ([]candidate_data.Candidate, error) {
	return uc.CandidateRepo.GetAllByType(candidateType)
}

// Vote votes for candidate by type, user_id, candidate_id.
func (uc *CandidateUseCase) Vote(candidateID uint, userID uint, candidateType candidate_data.CandidateType) error {
	if !candidate_data.IsValidCandidateType(string(candidateType)) {
		return errors.New("invalid candidate type")
	}

	voted, err := uc.VoteRepo.HasVoted(userID, string(candidateType))
	if err != nil {
		return err
	}
	if voted {
		return errors.New("already voted for this category")
	}

	candidate, err := uc.CandidateRepo.GetByID(candidateID)
	if err != nil {
		return err
	}

	if candidate.Type != candidateType {
		return errors.New("candidate type mismatch")
	}
	if time.Now().Before(candidate.VotingStart) {
		return errors.New("voting has not started for this candidate")
	}

	if time.Now().After(candidate.VotingDeadline) {
		return errors.New("voting period has ended for this candidate")
	}

	if err := uc.CandidateRepo.IncrementVote(candidateID); err != nil {
		return err
	}

	return uc.VoteRepo.SaveVote(candidateID, userID, string(candidateType))
}
