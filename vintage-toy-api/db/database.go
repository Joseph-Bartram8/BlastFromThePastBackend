package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	// Check if running locally (Render sets `RENDER` environment variable)
	_, isRender := os.LookupEnv("RENDER")

	// Load environment variables from .env file only if running locally
	if !isRender {
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠️ Warning: No .env file found, using system environment variables.")
		}
	}

	// Use DATABASE_URL if it exists (Render provides this)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// If DATABASE_URL isn't provided, fall back to manually building the connection string
		dbURL = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_SSLMODE"),
		)
	}

	// Connect to PostgreSQL
	var err error
	DB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("❌ Database connection failed:", err)
	}

	// Test connection
	err = DB.Ping()
	if err != nil {
		log.Fatal("❌ Database ping failed:", err)
	}

	fmt.Println("✅ Connected to the database!")
}
