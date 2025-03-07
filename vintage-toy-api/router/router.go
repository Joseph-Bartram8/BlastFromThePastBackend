package router

import (
	"database/sql"

	"github.com/Joseph_Bartram8/vintage-toy-api/handlers"
	"github.com/Joseph_Bartram8/vintage-toy-api/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// SetupRouter initializes the API routes
func SetupRouter(db *sql.DB) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.CORS)

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	// Public Routes
	r.Post("/login", handlers.LoginHandler(db))
	r.Post("/users", handlers.CreateUserHandler(db))
	r.Get("/users", handlers.GetUsersHandler(db))
	r.Post("/logout", handlers.LogoutHandler())
	r.Get("/markers", handlers.GetAllMarkersHandler(db))

	// Protected Routes
	r.Route("/api", func(api chi.Router) {
		api.Use(middleware.AuthMiddleware)

		api.Get("/user", handlers.GetCurrentUserHandler(db))
		api.Patch("/user", handlers.UpdateUserHandler(db))
		api.Delete("/user", handlers.DeleteUserHandler(db))
	})

	return r
}
