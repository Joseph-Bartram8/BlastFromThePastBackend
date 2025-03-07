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
	// Load environment variables (only needed for local development)
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No .env file found, using environment variables")
	}

	// Get connection string from environment variables
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("❌ DATABASE_URL environment variable is not set")
	}

	// Ensure Render-compatible connection settings (disable SSL for local but enable for production)
	if os.Getenv("ENV") != "production" {
		connStr += " sslmode=disable"
	}

	// Connect to PostgreSQL
	DB, err = sql.Open("postgres", connStr)
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
