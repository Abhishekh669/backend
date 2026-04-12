package jobs

import (
	"context"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/repository"
)

func StartAllOrderRelatedJobs(repo repository.OrderRepo) {
	go func() {
		ticker := time.NewTicker(20 * time.Minute)
		defer ticker.Stop()

		// Run once immediately on startup
		runOrderCleanupJobs(repo)

		for range ticker.C {
			runOrderCleanupJobs(repo)
		}
	}()
}

func runOrderCleanupJobs(repo repository.OrderRepo) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("🔄 [OrderJobs] Starting cleanup cycle...")

	if err := repo.DeleteInActiveTableValidation(ctx); err != nil {
		log.Printf("❌ [OrderJobs] DeleteInActiveTableValidation failed: %v", err)
	} else {
		log.Println("✅ [OrderJobs] DeleteInActiveTableValidation completed")
	}

	if err := repo.DeleteInApprovedOrders(ctx); err != nil {
		log.Printf("❌ [OrderJobs] DeleteInApprovedOrders failed: %v", err)
	} else {
		log.Println("✅ [OrderJobs] DeleteInApprovedOrders completed")
	}

	if err := repo.DeleteStaleTableSessions(ctx); err != nil {
		log.Printf("❌ [OrderJobs] DeleteStaleTableSessions failed: %v", err)
	} else {
		log.Println("✅ [OrderJobs] DeleteStaleTableSessions completed")
	}

	log.Println("🏁 [OrderJobs] Cleanup cycle done")
}
