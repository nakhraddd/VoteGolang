package candidate_usecase

import (
	"VoteGolang/internals/data"
	"VoteGolang/internals/repositories/candidate_repository"
	"VoteGolang/internals/repositories/votes_repositories"
	"errors"
)

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

func (uc *CandidateUseCase) GetAllByType(candidateType string) ([]data.Candidate, error) {
	return uc.CandidateRepo.GetAllByType(candidateType)
}

func (uc *CandidateUseCase) Vote(candidateID uint, userID uint, candidateType string) error {
	voted, err := uc.VoteRepo.HasVoted(userID, candidateType)
	if err != nil {
		return err
	}
	if voted {
		return errors.New("already voted for this category")
	}

	if err := uc.CandidateRepo.IncrementVote(candidateID); err != nil {
		return err
	}

	return uc.VoteRepo.SaveVote(candidateID, userID, candidateType)
}
