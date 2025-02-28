package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Abhishekh669/backend/configuration"
	"github.com/Abhishekh669/backend/middleware"
	"github.com/Abhishekh669/backend/router"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

var AdminPath = "/api/admin"

func main() {

	fmt.Println("Starting server...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database and Google OAuth
	configuration.InitDataBase()
	configuration.InitGoogleAuthConfig()

	// Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Cookie", "Authorization"},
		AllowCredentials: true,
	})
	r := mux.NewRouter()

	//socket io
	// Define excluded paths
	excludedPaths := []string{
		"/api/auth/google",
		"/api/auth/fuck",
		"/api/auth/google/callback",
		"/api/user",
		"/api/auth/session",
		"/api/ws",
	}

	// Create Gorilla Mux router

	// Apply middleware
	r.Use(ApplyMiddlewareExcept(excludedPaths...))

	// Mount sub-routers
	r.PathPrefix("/api/user").Handler(router.UserRouter())
	r.PathPrefix("/api/auth").Handler(router.AuthRouter())
	r.PathPrefix("/api/chat").Handler(router.ChatRouter())
	r.PathPrefix("/api/message").Handler(router.MessageRouter())
	r.PathPrefix("/api/admin").Handler(router.AdminRouter())
	r.PathPrefix("/api/timer").Handler(router.TimerRouter())
	r.PathPrefix("/api/ws").Handler(router.SocketRouter())
	r.PathPrefix("/api/student/absent").Handler(router.StudentAbsentRouter())

	// Start server with graceful shutdown
	server := &http.Server{
		Addr:    ":8000",
		Handler: c.Handler(r),
	}

	go func() {
		log.Println("Server starting on port :8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// ApplyMiddlewareExcept applies middleware to all routes except the excluded paths
func ApplyMiddlewareExcept(excludedPaths ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip middleware for excluded paths
			for _, path := range excludedPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Admin validation for admin routes
			if strings.HasPrefix(r.URL.Path, AdminPath) && !isExcludedAdminPath(r.URL.Path) {
				if !middleware.ValidateAdmin(w, r) {
					http.Error(w, "Admin Forbidden", http.StatusForbidden)
					return
				}
			}

			// Session validation for all other routes
			if !middleware.ValidateSession(w, r) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isExcludedAdminPath checks if the admin path is excluded from validation
func isExcludedAdminPath(path string) bool {
	excludedAdminPaths := []string{
		"/api/admin/new",
		"/api/admin/login",
		"/api/admin/all",
	}

	for _, p := range excludedAdminPaths {
		if path == p {
			return true
		}
	}

	return false
}
