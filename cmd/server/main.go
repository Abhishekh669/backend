package main

import (
	"context"
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

	app, err := app.New()

	if err != nil {
		log.Fatalf("❌ Failed to initialize app: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		// AllowOrigins: []string{"http://localhost:3000", "https://kitbmantra.vercel.app"},
		// AllowAllOrigins: "",
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

	newCache := algorithm.NewMenuCache()
	foodCategories, err := app.FoodCategoryRepo.GetAllCategoriesFromDB(context.Background())
	if err != nil {
		log.Println("Error in getting food categories")

	}

	menuItems, err := app.FoodCategoryRepo.GetAllMenuItemsFromDB(context.Background())
	if err != nil {
		log.Println("error in getting menu times ")

	}
	newCache.ReloadFromDB(foodCategories, menuItems)
	router.GET("/get-menu-n-categories", func(c *gin.Context) {

		categories, categoryChildren, menuItems :=
			newCache.GetFullMenuSnapshot()

		c.JSON(http.StatusOK, gin.H{
			"success":           true,
			"categories":        categories,
			"category_children": categoryChildren,
			"menu_items":        menuItems,
		})
	})
	routes.SetUpRoutes(router, app, newCache)
	jobs.StartDailyAttendanceReview(app.AttendanceRepo)

	log.Println("🌐 Starting HTTP server on :8080...")
	log.Println("📡 Server endpoints are now available")
	log.Println("App ready in", time.Since(start).Seconds(), "seconds.")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}

}
