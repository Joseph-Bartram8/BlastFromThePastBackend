package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/Joseph_Bartram8/vintage-toy-api/db"
	"github.com/Joseph_Bartram8/vintage-toy-api/router"
)

func main() {
	if os.Getenv("DATABASE_URL") == "" {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ Warning: No .env file found, using system environment variables")
		}
	}

	// Connect to database
	db.ConnectDB()

	// Initialize router with database instance
	r := router.SetupRouter(db.DB)

	// Determine port for local or Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to 8080 for local testing
	}

	// Start server
	log.Printf("🚀 Server running on :%s\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}
