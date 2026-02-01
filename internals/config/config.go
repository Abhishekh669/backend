package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgressURL  string
	FrontendURL   string
	JWTToken      string
	AdminEmail    string
	AdminPassword string
	AdminPhone    string
}

var AppConfig *Config

func InitConfig() error {
	if err := godotenv.Load(); err != nil {
		log.Printf("Waring : Couldnot load .env files : %v", err)
		return fmt.Errorf("couldn't load .env files : %v", err)
	}

	AppConfig = &Config{
		PostgressURL:  os.Getenv("DATABASE_URL"),
		FrontendURL:   os.Getenv("PUBLIC_FRONTEND_URL"),
		JWTToken:      os.Getenv("JWT_TOKEN"),
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		AdminPhone:    os.Getenv("ADMIN_PHONE"),
	}

	return nil
}
