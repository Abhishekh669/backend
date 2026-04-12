package jobs

import (
	"context"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/repository"
)

func StartTokenCleanupJob(repo repository.PaymentRepo) {
	go func() {
		ticker := time.NewTicker(30 * 24 * time.Hour)
		defer ticker.Stop()

		runTokenCleanupJobs(repo)

		for range ticker.C {
			runTokenCleanupJobs(repo)
		}
	}()
}

func runTokenCleanupJobs(repo repository.PaymentRepo) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("🔄 [TokenJobs] Starting token cleanup cycle...")

	if err := repo.ResetExpiredUserTokens(ctx); err != nil {
		log.Printf("❌ [TokenJobs] ResetExpiredUserTokens failed: %v", err)
	} else {
		log.Println("✅ [TokenJobs] ResetExpiredUserTokens completed")
	}

	log.Println("🏁 [TokenJobs] Token cleanup cycle done")
}
