package petition_usecase

import (
	"VoteGolang/internals/app/logging"
	"VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/repositories"
	"VoteGolang/internals/service"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

// PetitionUseCase manages petition creation and retrieval.
type PetitionUseCase interface {
	CreatePetition(p *domain.Petition) error
	GetAllPetitions() ([]domain.Petition, error)
	GetPetitionByID(id uint) (*domain.Petition, error)
	Vote(userID uint, petitionID uint, voteType domain.VoteType) error
	DeletePetition(id uint) error
	HasUserVoted(userID uint, petitionID uint) (bool, error)
	GetAllPetitionsPaginated(limit, offset int) ([]domain.Petition, error)
}

type petitionUseCase struct {
	petitionRepo     domain.PetitionRepository
	petitionVoteRepo domain.PetitionVoteRepository
	blockchain       service.BlockchainService
	redis            *redis.Client
	logger           *logging.KafkaLogger
	searchRepo       *repositories.SearchRepository
}

// NewPetitionUseCase updated to include KafkaLogger
func NewPetitionUseCase(
	pr domain.PetitionRepository,
	pvr domain.PetitionVoteRepository,
	bc service.BlockchainService,
	rdb *redis.Client,
	kafkaLogger *logging.KafkaLogger,
	searchRepo *repositories.SearchRepository,
) PetitionUseCase {
	return &petitionUseCase{
		petitionRepo:     pr,
		petitionVoteRepo: pvr,
		blockchain:       bc,
		redis:            rdb,
		logger:           kafkaLogger,
		searchRepo:       searchRepo,
	}
}

func (uc *petitionUseCase) GetAllPetitionsPaginated(limit, offset int) ([]domain.Petition, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("petitions:page:%d:limit:%d", offset/limit+1, limit)

	cached, err := uc.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var petitions []domain.Petition
		if err := json.Unmarshal([]byte(cached), &petitions); err == nil {
			uc.logger.Log("DEBUG", fmt.Sprintf("Cache hit for: %s", cacheKey))
			return petitions, nil
		}
	}

	uc.logger.Log("DEBUG", fmt.Sprintf("Cache miss for: %s", cacheKey))

	petitions, err := uc.petitionRepo.GetAllPaginated(limit, offset)
	if err != nil {
		uc.logger.Log("ERROR", fmt.Sprintf("Failed to get paginated petitions from DB: %v", err))
		return nil, err
	}

	bytes, _ := json.Marshal(petitions)
	uc.redis.Set(ctx, cacheKey, bytes, 10*time.Minute)
	return petitions, nil
}

func (uc *petitionUseCase) CreatePetition(p *domain.Petition) error {
	if err := uc.petitionRepo.Create(p); err != nil {
		uc.logger.Log("ERROR", fmt.Sprintf("Failed to create petition in DB: %v", err))
		return err
	}
	uc.logger.Log("INFO", fmt.Sprintf("Petition %d created successfully in DB", p.ID))

	if uc.searchRepo != nil {
		go func() {
			id := fmt.Sprintf("%d", p.ID)
			if err := uc.searchRepo.Index(context.Background(), id, p); err != nil {
				uc.logger.Log("WARN", fmt.Sprintf("Failed to index petition %d: %v", p.ID, err))
			} else {
				uc.logger.Log("DEBUG", fmt.Sprintf("Petition %d indexed for search", p.ID))
			}
		}()
	}

	// Invalidate cache
	uc.invalidateAllPetitionCaches()

	// Log to blockchain
	if _, err := uc.blockchain.LogPetitionCreation(p); err != nil {
		uc.logger.Log("ERROR", fmt.Sprintf("CRITICAL: Petition %d created in DB but failed to log to blockchain: %v", p.ID, err))
		// Do not return error, as the petition *was* created.
	} else {
		uc.logger.Log("INFO", fmt.Sprintf("Petition %d logged to blockchain", p.ID))
	}

	return nil
}

func (uc *petitionUseCase) GetAllPetitions() ([]domain.Petition, error) {
	ctx := context.Background()
	cacheKey := "petitions"

	cached, err := uc.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var petitions []domain.Petition
		if err := json.Unmarshal([]byte(cached), &petitions); err == nil {
			uc.logger.Log("DEBUG", fmt.Sprintf("Cache hit for: %s", cacheKey))
			return petitions, nil
		}
	}

	uc.logger.Log("DEBUG", fmt.Sprintf("Cache miss for: %s", cacheKey))
	petitions, err := uc.petitionRepo.GetAll()
	if err != nil {
		uc.logger.Log("ERROR", fmt.Sprintf("Failed to get all petitions from DB: %v", err))
		return nil, err
	}

	data, _ := json.Marshal(petitions)
	uc.redis.Set(ctx, cacheKey, data, time.Duration(rand.Intn(5)+25)*time.Minute)
	return petitions, nil
}

func (uc *petitionUseCase) GetPetitionByID(id uint) (*domain.Petition, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("petition:%d", id)

	if cached, err := uc.redis.Get(ctx, cacheKey).Result(); err == nil {
		var petition domain.Petition
		if json.Unmarshal([]byte(cached), &petition) == nil {
			uc.logger.Log("DEBUG", fmt.Sprintf("Cache hit for: %s", cacheKey))
			return &petition, nil
		}
	}

	uc.logger.Log("DEBUG", fmt.Sprintf("Cache miss for: %s", cacheKey))
	petition, err := uc.petitionRepo.GetByID(id)
	if err != nil {
		uc.logger.Log("WARN", fmt.Sprintf("Failed to get petition %d from DB: %v", id, err))
		return nil, err
	}

	data, _ := json.Marshal(petition)
	uc.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	return petition, nil
}

func (uc *petitionUseCase) Vote(userID uint, petitionID uint, voteType domain.VoteType) error {
	// 1. Pre-flight checks
	voted, err := uc.petitionVoteRepo.HasUserVoted(userID, petitionID)
	if err != nil {
		uc.logger.Log("ERROR", fmt.Sprintf("Failed to check HasUserVoted (user: %d, petition: %d): %v", userID, petitionID, err))
		return err
	}
	if voted {
		return fmt.Errorf("user has already voted")
	}

	petition, err := uc.petitionRepo.GetByID(petitionID)
	if err != nil {
		uc.logger.Log("WARN", fmt.Sprintf("Vote attempt on non-existent petition %d: %v", petitionID, err))
		return err
	}

	if !domain.IsValidVoteType(string(voteType)) {
		return fmt.Errorf("invalid petition type: must be 'favor' or 'against'")
	}
	if time.Now().After(petition.VotingDeadline) {
		return fmt.Errorf("voting period has ended")
	}
	totalVotes := petition.VotesInFavor + petition.VotesAgainst
	if totalVotes >= petition.Goal {
		return fmt.Errorf("petition goal has been reached")
	}

	// This callback contains *only* the DB logic.
	dbTransactionCallback := func() error {
		var dbErr error
		switch voteType {
		case domain.Favor:
			dbErr = uc.petitionRepo.VoteInFavor(petitionID)
		case domain.Against:
			dbErr = uc.petitionRepo.VoteAgainst(petitionID)
		default:
			return fmt.Errorf("invalid vote type") // Should be caught by pre-flight, but good to double check
		}
		return dbErr
	}

	// VoteWithTransaction will Begin, execute the callback, save the vote record, and Commit/Rollback
	err = uc.petitionVoteRepo.VoteWithTransaction(userID, petitionID, voteType, dbTransactionCallback)
	if err != nil {
		uc.logger.Log("ERROR", fmt.Sprintf("DB transaction for vote failed (user: %d, petition: %d): %v", userID, petitionID, err))
		return fmt.Errorf("database transaction failed: %w", err)
	}

	if _, err := uc.blockchain.LogPetitionVote(userID, petitionID, voteType); err != nil {
		uc.logger.Log("ERROR", fmt.Sprintf("CRITICAL: Petition vote (user %d, petition %d) saved to DB but failed to log to blockchain: %v", userID, petitionID, err))
		// Do not return error, the vote was successful in the DB.
	} else {
		uc.logger.Log("INFO", fmt.Sprintf("Petition vote (user %d, petition %d) logged to blockchain", userID, petitionID))
	}

	// We can invalidate both the specific petition and the paginated lists.
	ctx := context.Background()
	cacheKey := fmt.Sprintf("petition:%d", petitionID)
	uc.redis.Del(ctx, cacheKey)
	uc.invalidateAllPetitionCaches()

	uc.logger.Log("INFO", fmt.Sprintf("Vote cast by user %d on petition %d", userID, petitionID))
	return nil
}

func (uc *petitionUseCase) DeletePetition(id uint) error {
	if err := uc.petitionRepo.Delete(id); err != nil {
		uc.logger.Log("ERROR", fmt.Sprintf("Failed to delete petition %d: %v", id, err))
		return err
	}

	// Invalidate all petition caches
	uc.invalidateAllPetitionCaches()
	uc.logger.Log("INFO", fmt.Sprintf("Petition %d deleted and cache invalidated", id))
	return nil
}

func (uc *petitionUseCase) HasUserVoted(userID uint, petitionID uint) (bool, error) {
	return uc.petitionVoteRepo.HasUserVoted(userID, petitionID)
}

// invalidateAllPetitionCaches is a helper to clear list/paginated caches
func (uc *petitionUseCase) invalidateAllPetitionCaches() {
	ctx := context.Background()
	pattern := "petitions*"
	var cursor uint64
	var keysFound int

	for {
		keys, nextCursor, err := uc.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			uc.logger.Log("WARN", fmt.Sprintf("Failed to scan Redis keys for pattern %s: %v", pattern, err))
			return
		}

		if len(keys) > 0 {
			uc.redis.Del(ctx, keys...)
			keysFound += len(keys)
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	uc.logger.Log("DEBUG", fmt.Sprintf("Cache invalidated for pattern '%s' (keys deleted: %d)", pattern, keysFound))
}
