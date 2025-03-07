package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/rs/cors" // Import CORS package

	"github.com/Joseph_Bartram8/vintage-toy-api/db"
	"github.com/Joseph_Bartram8/vintage-toy-api/router"
)

func main() {
	// Connect to database
	db.ConnectDB()

	// Initialize router with database instance
	r := router.SetupRouter(db.DB)

	// Set up CORS middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://blastfromthepastbackend.onrender.com", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// Get port from environment variable (Render provides this dynamically)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to 8080 if not set (for local testing)
	}

	// Start server with CORS handling
	log.Printf("🚀 Server running on :%s\n", port)
	err := http.ListenAndServe(":"+port, corsHandler.Handler(r))
	if err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}
