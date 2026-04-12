package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Abhishekh669/backend/internals/algorithm"
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/config"
	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/jobs"
	"github.com/Abhishekh669/backend/internals/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	start := time.Now()

	if err := config.InitConfig(); err != nil {
		log.Printf("%v", err)
	}

	if err := database.InitializeDatabase(); err != nil {
		log.Printf("db conn err : %v", err)
	}

	application, err := app.New()
	if err != nil {
		log.Fatalf("❌ Failed to initialize app: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "*", "https://rms-gules.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-App-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.HEAD("/", func(c *gin.Context) {
		c.Status(200)
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"pong": "pong"})
	})

	// ── Menu cache (existing) ──────────────────────────────────────────────────
	newCache := algorithm.NewMenuCache(application.FoodCategoryRepo)
	newCache.ReloadFromDB()

	router.GET("/get-menu-n-categories", func(c *gin.Context) {
		groupedMenu := newCache.GetAll()
		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"grouped_menu": groupedMenu,
		})
	})

	// ── Report cache: load on startup in background ────────────────────────────
	// Non-blocking: server starts immediately, cache fills in background.
	// The /admin/reports/status endpoint tells the frontend when data is ready.
	go func() {
		log.Println("📊 [ReportCache] Loading initial report data in background...")
		application.ReportCache.ReloadFromDB()
	}()

	// ── Nightly report refresh job (00:05 every night) ─────────────────────────
	jobs.StartNightlyReportRefresh(application.ReportCache)
	jobs.StartAllOrderRelatedJobs(application.OrderRepo)
	jobs.StartTokenCleanupJob(application.PaymentRepo)
	// ── Existing jobs ──────────────────────────────────────────────────────────
	jobs.StartDailyAttendanceReview(application.AttendanceRepo)

	routes.SetUpRoutes(router, application, newCache)

	log.Println("🌐 Starting HTTP server on :8080...")
	log.Println("📡 Server endpoints are now available")
	log.Println("App ready in", time.Since(start).Seconds(), "seconds.")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
