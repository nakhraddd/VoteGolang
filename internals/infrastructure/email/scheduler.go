package email

import (
	"VoteGolang/internals/domain"
	"log"
	"time"
)

func StartUnverifiedCleanupJob(repo domain.UserRepository) {
	//ticker := time.NewTicker(10 * time.Second)
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			cutoff := time.Now().Add(-24 * time.Hour)
			//cutoff := time.Now().Add(-10 * time.Second)
			deleted, err := repo.DeleteUnverifiedUser(cutoff)
			if err != nil {
				log.Println("Cleanup error:", err)
			} else if deleted > 0 {
				log.Printf("Deleted %d unverified users\n", deleted)
			}
		}
	}()
}
