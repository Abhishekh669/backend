package jobs

import (
	"context"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/algorithm"
)

// StartAssociationRulesRefresh runs a background goroutine that refreshes
// recommendation association rules at a fixed interval (e.g. every 48 hours).
func StartAssociationRulesRefresh(cache *algorithm.CacheManager, interval time.Duration) {
	if cache == nil {
		log.Println("⚠️ [AssociationRulesRefresh] cache is nil; job not started")
		return
	}
	if interval <= 0 {
		interval = 48 * time.Hour
	}

	go func() {
		log.Printf("🔁 [AssociationRulesRefresh] Job started — refresh every %s", interval)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("🔄 [AssociationRulesRefresh] Refreshing association rules cache...")
			if err := cache.Refresh(context.Background()); err != nil {
				log.Printf("❌ [AssociationRulesRefresh] refresh failed: %v", err)
				continue
			}
			log.Println("✅ [AssociationRulesRefresh] refresh completed")
		}
	}()
}
