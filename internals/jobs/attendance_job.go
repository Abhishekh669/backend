package jobs

import (
	"context"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/repository"
)

func StartDailyAttendanceReview(repo repository.AttendanceRepo) {
	go func() {
		for {
			now := time.Now()

			next := time.Date(
				now.Year(),
				now.Month(),
				now.Day(),
				0, 5, 0, 0,
				now.Location(),
			)

			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}

			time.Sleep(time.Until(next))

			log.Println("🕛 Running daily attendance auto-review job...")

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			err := repo.AutoReviewIncompleteAttendance(ctx)
			if err != nil {
				log.Println("❌ Attendance auto-review failed:", err)
			} else {
				log.Println("✅ Attendance auto-review completed successfully")
			}
			cancel()
		}
	}()
}
