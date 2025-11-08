package candidate_usecase

import (
	"VoteGolang/internals/blockchain"
	candidate_data2 "VoteGolang/internals/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"VoteGolang/internals/infrastructure/search"
	"github.com/redis/go-redis/v9"
)

// CandidateUseCase handles business logic related to election candidates.
type CandidateUseCase struct {
	CandidateRepo candidate_data2.CandidateRepository
	VoteRepo      candidate_data2.VoteRepository
	Blockchain    *blockchain.Blockchain
	Redis         *redis.Client
	SearchRepo    *search.SearchRepository
}

func NewCandidateUseCase(cRepo candidate_data2.CandidateRepository, vRepo candidate_data2.VoteRepository, bc *blockchain.Blockchain, rdb *redis.Client, searchRepo *search.SearchRepository) *CandidateUseCase {
	return &CandidateUseCase{
		CandidateRepo: cRepo,
		VoteRepo:      vRepo,
		Blockchain:    bc,
		Redis:         rdb,
		SearchRepo:    searchRepo,
	}
}
func (uc *CandidateUseCase) CreateCandidate(candidate *candidate_data2.Candidate) error {
	if err := uc.CandidateRepo.Create(candidate); err != nil {
		return err
	}

	// Index in Elasticsearch
	if uc.SearchRepo != nil {
		go func() {
			id := fmt.Sprintf("%d", candidate.ID)
			if err := uc.SearchRepo.IndexDocument(id, candidate); err != nil {
				log.Printf("❌ Failed to index candidate: %v", err)
			}
		}()
	}

	// Invalidate cache
	pattern := fmt.Sprintf("candidates:type:%s*", candidate.Type)
	keys, _ := uc.Redis.Keys(context.Background(), pattern).Result()
	for _, k := range keys {
		uc.Redis.Del(context.Background(), k)
	}

	transaction := blockchain.Transaction{
		Type:    "CANDIDATE_CREATION",
		Payload: candidate,
	}
	uc.Blockchain.AddBlock(transaction)
	return nil
}

func (uc *CandidateUseCase) GetAllByTypePaginated(candidateType string, limit, offset int) ([]candidate_data2.Candidate, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("candidates:type:%s:page:%d:limit:%d", candidateType, offset/limit+1, limit)

	cached, err := uc.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var candidates []candidate_data2.Candidate
		if err := json.Unmarshal([]byte(cached), &candidates); err == nil {
			log.Println("Cache hit:", cacheKey)
			return candidates, nil
		}
	}

	log.Println("Cache miss:", cacheKey)

	candidates, err := uc.CandidateRepo.GetAllByTypePaginated(candidateType, limit, offset)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(candidates)
	uc.Redis.Set(ctx, cacheKey, data, time.Duration(rand.Intn(5)+25)*time.Minute) // 25–30 minutes

	return candidates, nil
}

// GetAllByType returns a list of candidates filtered by type.
func (uc *CandidateUseCase) GetAllByType(candidateType string) ([]candidate_data2.Candidate, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("candidates:type:%s", candidateType)

	// Try cache first
	cached, err := uc.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var candidates []candidate_data2.Candidate
		if err := json.Unmarshal([]byte(cached), &candidates); err == nil {
			log.Println("Cache hit:", cacheKey)
			return candidates, nil
		}
	}

	log.Println("Cache miss:", cacheKey)
	// Fallback to DB
	candidates, err := uc.CandidateRepo.GetAllByType(candidateType)
	if err != nil {
		return nil, err
	}

	// Save to Redis
	data, _ := json.Marshal(candidates)
	uc.Redis.Set(ctx, cacheKey, data, time.Duration(rand.Intn(5)+25)*time.Minute) // 25–30 minutes
	return candidates, nil
}

// Vote votes for candidate by type, user_id, candidate_id.
func (uc *CandidateUseCase) Vote(candidateID uint, userID uint, candidateType candidate_data2.CandidateType) error {
	if !candidate_data2.IsValidCandidateType(string(candidateType)) {
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

	if err := uc.VoteRepo.SaveVote(candidateID, userID, string(candidateType)); err != nil {
		return err
	}

	transaction := blockchain.Transaction{
		Type: "VOTE_CAST",
		Payload: map[string]interface{}{
			"candidate_id":   candidateID,
			"user_id":        userID,
			"candidate_type": candidateType,
		},
	}
	uc.Blockchain.AddBlock(transaction)

	return nil
}

func (uc *CandidateUseCase) DeleteCandidate(id uint) error {
	if err := uc.CandidateRepo.DeleteByID(id); err != nil {
		return err
	}

	// Invalidate all petition caches
	ctx := context.Background()
	var cursor uint64
	for {
		keys, nextCursor, _ := uc.Redis.Scan(ctx, cursor, "candidates*", 100).Result()
		for _, k := range keys {
			uc.Redis.Del(ctx, k)
		}
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return nil
}
