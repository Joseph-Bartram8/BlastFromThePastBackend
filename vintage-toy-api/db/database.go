package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	var dbURI string

	// Check if USE_RENDER_DB is set to "true"
	useRenderDB, _ := strconv.ParseBool(os.Getenv("USE_RENDER_DB"))

	if useRenderDB {
		// Use Render database
		dbURI = os.Getenv("DATABASE_URL")
		fmt.Println("🌍 Using Render Database")
	} else {
		// Use Local Database
		fmt.Println("💻 Using Local Database")
		dbURI = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_SSLMODE"),
		)
	}

	// Log database connection string (excluding password for security)
	fmt.Println("🔍 Database connection string:", dbURI)

	// Connect to the database
	var err error
	DB, err = sql.Open("postgres", dbURI)
	if err != nil {
		log.Fatalf("❌ Failed to open database: %v", err)
	}

	// Verify connection
	if err = DB.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}

	fmt.Println("✅ Successfully connected to database!")
}
