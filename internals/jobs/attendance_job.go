package jobs

import (
	"context"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/repository"
)

func StartDailyAttendanceReview(repo repository.AttendanceRepo) {
	go func() {
		loc, _ := time.LoadLocation("Asia/Kathmandu")

		for {
			now := time.Now().In(loc)

			next := time.Date(
				now.Year(),
				now.Month(),
				now.Day(),
				0, 5, 0, 0,
				loc,
			)

			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}

			time.Sleep(time.Until(next))

			log.Println("🕛 Running daily attendance jobs...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

			// 🔥 STEP 1: Create daily attendance (MOST IMPORTANT)
			err := repo.CreateDailyAbsentAttendance(ctx)
			if err != nil {
				log.Println("❌ Create daily attendance failed:", err)
			} else {
				log.Println("✅ Daily attendance created successfully")
			}

			// 🔥 STEP 2: Auto review (your existing logic)
			err = repo.AutoReviewIncompleteAttendance(ctx)
			if err != nil {
				log.Println("❌ Attendance auto-review failed:", err)
			} else {
				log.Println("✅ Attendance auto-review completed successfully")
			}

			// 🔥 STEP 3: Cleanup leave requests
			err = repo.DeleteInactiveLeaveRequestAttendance(ctx)
			if err != nil {
				log.Println("❌ Deleting inactive leave request attendance failed:", err)
			} else {
				log.Println("✅ Leave cleanup completed successfully")
			}

			cancel()
		}
	}()
}
