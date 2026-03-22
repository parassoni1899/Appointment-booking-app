package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl string
	Port  string
}

func LoadConfig() *Config {
	// Attempt to load .env, ignore if not found
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		// fallback to a default local postgres url for ease of setup
		dbUrl = "host=localhost user=postgres password=postgres dbname=booking port=5432 sslmode=disable TimeZone=UTC"
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DBUrl: dbUrl,
		Port:  port,
	}
}
