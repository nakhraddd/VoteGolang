package petittion_usecase

import (
	"VoteGolang/internals/domain"  // This was petition_data2, aliasing to domain
	"VoteGolang/internals/service" // <-- NEW IMPORT
	"fmt"
	"log"
	"math/rand"
	"time"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

// PetitionUseCase manages petition creation and retrieval.
type PetitionUseCase interface {
	// CreatePetition allows a user to create a new petition.
	CreatePetition(p *domain.Petition) error
	// GetAllPetitions returns all active petitions.
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
	blockchain       service.BlockchainService // <-- CHANGED
	redis            *redis.Client
}

func NewPetitionUseCase(pr petition_data2.PetitionRepository, pvr petition_data2.PetitionVoteRepository, bc *blockchain.Blockchain, rdb *redis.Client) PetitionUseCase {
	return &petitionUseCase{
		petitionRepo:     pr,
		petitionVoteRepo: pvr,
		blockchain:       bc,
		redis:            rdb,
	}
}
func (uc *petitionUseCase) GetAllPetitionsPaginated(limit, offset int) ([]petition_data2.Petition, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("petitions:page:%d:limit:%d", offset/limit+1, limit)

	cached, err := uc.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var petitions []petition_data2.Petition
		if err := json.Unmarshal([]byte(cached), &petitions); err == nil {
			log.Println("Cache hit:", cacheKey)
			return petitions, nil
		}
	}

	log.Println("Cache miss:", cacheKey)

	petitions, err := uc.petitionRepo.GetAllPaginated(limit, offset)
	if err != nil {
		return nil, err
	}

	bytes, _ := json.Marshal(petitions)
	uc.redis.Set(ctx, cacheKey, bytes, 10*time.Minute)
	return petitions, nil
}

func (uc *petitionUseCase) CreatePetition(p *domain.Petition) error {
	if err := uc.petitionRepo.Create(p); err != nil {
		return err
	}
	// Invalidate cache
	pattern := "petitions*"
	ctx := context.Background()
	var cursor uint64
	for {
		keys, nextCursor, _ := uc.redis.Scan(ctx, cursor, pattern, 100).Result()
		for _, k := range keys {
			uc.redis.Del(ctx, k)
		}
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	if _, err := uc.blockchain.LogPetitionCreation(p); err != nil {
		log.Printf("ERROR: Petition %d created in DB but failed to log to blockchain: %v", p.ID, err)
		// See note in candidate_usecase about handling this error
	}

	return nil
}

func (uc *petitionUseCase) GetAllPetitions() ([]domain.Petition, error) {
	ctx := context.Background()
	cacheKey := "petitions"

	// Try cache first
	cached, err := uc.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var petitions []petition_data2.Petition
		if err := json.Unmarshal([]byte(cached), &petitions); err == nil {
			log.Println("Cache hit:", cacheKey)
			return petitions, nil
		}
	}

	log.Println("Cache miss:", cacheKey)
	// Fallback to DB
	petitions, err := uc.petitionRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// Save to Redis
	data, _ := json.Marshal(petitions)
	uc.redis.Set(ctx, cacheKey, data, time.Duration(rand.Intn(5)+25)*time.Minute) // 25–30 minutes
	return petitions, nil
}

func (uc *petitionUseCase) GetPetitionByID(id uint) (*domain.Petition, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("petition:%d", id)

	if cached, err := uc.redis.Get(ctx, cacheKey).Result(); err == nil {
		var petition petition_data2.Petition
		if json.Unmarshal([]byte(cached), &petition) == nil {
			log.Println("Cache hit:", cacheKey)
			return &petition, nil
		}
	}

	log.Println("Cache miss:", cacheKey)
	petition, err := uc.petitionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(petition)
	uc.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	return petition, nil
}

func (uc *petitionUseCase) Vote(userID uint, petitionID uint, voteType domain.VoteType) error {
	voted, err := uc.petitionVoteRepo.HasUserVoted(userID, petitionID)
	if err != nil {
		return err
	}
	if voted {
		return fmt.Errorf("user has already voted")
	}

	petition, err := uc.petitionRepo.GetByID(petitionID)
	if err != nil {
		return err
	}

	if !petition_data2.IsValidVoteType(string(voteType)) {
		return fmt.Errorf("invalid petition type: must be 'favor' or 'against'")
	}

	if time.Now().After(petition.VotingDeadline) {
		return fmt.Errorf("voting period has ended")
	}

	totalVotes := petition.VotesInFavor + petition.VotesAgainst
	if totalVotes >= petition.Goal {
		return fmt.Errorf("petition goal has been reached")
	}

	// Use transaction with row locking for idempotency
	return uc.petitionVoteRepo.VoteWithTransaction(userID, petitionID, voteType, func() error {
		// Update vote count based on vote type
		var dbErr error
		switch voteType {
		case petition_data2.Favor:
			dbErr = uc.petitionRepo.VoteInFavor(petitionID)
		case petition_data2.Against:
			dbErr = uc.petitionRepo.VoteAgainst(petitionID)
		default:
			return fmt.Errorf("invalid vote type")
		}

		if dbErr != nil {
			return dbErr
		}

		// Add to blockchain
		go func() {
			if uc.blockchain != nil {
				transaction := blockchain.Transaction{
					Type: "PETITION_VOTE",
					Payload: map[string]interface{}{
						"petition_id": petitionID,
						"user_id":     userID,
						"vote_type":   voteType,
					},
					Description: fmt.Sprintf("User %d voted on petition %d", userID, petitionID),
					Timestamp:   time.Now(),
				}
				if _, err := uc.blockchain.LogPetitionVote(userID, petitionID, voteType); err != nil {
					log.Printf("ERROR: Petition vote (user %d, petition %d) saved to DB but failed to log to blockchain: %v", userID, petitionID, err)
				}			}
		}()
		// Invalidate cache
		ctx := context.Background()
		cacheKey := fmt.Sprintf("petition:%d", petitionID)
		uc.redis.Del(ctx, cacheKey)

		// Invalidate paginated cache
		var cursor uint64
		for {
			keys, nextCursor, _ := uc.redis.Scan(ctx, cursor, "petitions*", 100).Result()
			for _, k := range keys {
				uc.redis.Del(ctx, k)
			}
			if nextCursor == 0 {
				break
			}
			cursor = nextCursor
		}

		return nil
	})
}

func (uc *petitionUseCase) DeletePetition(id uint) error {
	if err := uc.petitionRepo.Delete(id); err != nil {
		return err
	}

	// Invalidate all petition caches
	ctx := context.Background()
	var cursor uint64
	for {
		keys, nextCursor, _ := uc.redis.Scan(ctx, cursor, "petitions*", 100).Result()
		for _, k := range keys {
			uc.redis.Del(ctx, k)
		}
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return nil
}

func (uc *petitionUseCase) HasUserVoted(userID uint, petitionID uint) (bool, error) {
	return uc.petitionVoteRepo.HasUserVoted(userID, petitionID)
}
