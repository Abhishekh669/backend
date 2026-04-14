package jobs

import (
	"context"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/repository"
)

func StartForgetPasswordCleanupJob(repo repository.UserRepo) {
	go func() {
		ticker := time.NewTicker(30 * 24 * time.Hour)
		defer ticker.Stop()

		runForgetPasswordCleanupJobs(repo)

		for range ticker.C {
			runForgetPasswordCleanupJobs(repo)
		}
	}()
}

func runForgetPasswordCleanupJobs(repo repository.UserRepo) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("🔄 [ForgetPasswordJobs] Starting forget password cleanup cycle...")

	if err := repo.CleanupExpiredNUsedForgetPasswordSessions(ctx); err != nil {
		log.Printf("❌ [ForgetPasswordJobs] CleanupExpiredNUsedForgetPasswordSessions failed: %v", err)
	} else {
		log.Println("✅ [ForgetPasswordJobs] CleanupExpiredNUsedForgetPasswordSessions completed")
	}

	log.Println("🏁 [ForgetPasswordJobs] Forget password cleanup cycle done")
}
