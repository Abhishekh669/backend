package jobs

import (
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/algorithm"
)

// StartNightlyReportRefresh runs every night at 00:05 (just after midnight)
// to pre-compute and cache all report data.
// It follows the same pattern as StartDailyAttendanceReview.
func StartNightlyReportRefresh(reportCache *algorithm.DefaultRevenueCache) {
	go func() {
		log.Println("🌙 [NightlyReportRefresh] Job started — will refresh reports at 00:05 every night")

		for {
			now := time.Now()

			// Calculate next 00:05
			next := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, now.Location())
			if now.After(next) {
				// Already past 00:05 today — schedule for tomorrow
				next = next.Add(24 * time.Hour)
			}

			waitDuration := time.Until(next)
			log.Printf("🕛 [NightlyReportRefresh] Next refresh scheduled in %s (at %s)",
				waitDuration.Round(time.Second),
				next.Format("2006-01-02 15:04:05"),
			)

			time.Sleep(waitDuration)

			log.Println("🌙 [NightlyReportRefresh] Running nightly report cache refresh...")
			reportCache.ReloadFromDB()
		}
	}()
}
