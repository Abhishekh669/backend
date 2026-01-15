package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.HEAD("/", func(c *gin.Context) {
		c.Status(200)
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"pong": "pong"})
	})

	router.Use(cors.New(cors.Config{
		// AllowOrigins: []string{"http://localhost:3000", "https://kitbmantra.vercel.app"},
		// AllowAllOrigins: "",
		AllowOrigins:     []string{"http://localhost:3000", "*", "https://rms-gules.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-App-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	log.Println("🌐 Starting HTTP server on :8080...")
	log.Println("📡 Server endpoints are now available")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}

}
