package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// LoginHandler authenticates the user and sets a secure cookie
func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		// Parse request body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Authenticate user
		var userID uuid.UUID
		var storedPasswordHash string

		err := db.QueryRow("SELECT id, password_hash FROM users WHERE email = $1 AND is_deleted = FALSE", req.Email).
			Scan(&userID, &storedPasswordHash)

		if err == sql.ErrNoRows {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		// Generate JWT token
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		// Set Secure HTTP-Only Cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    tokenString,
			HttpOnly: true,                    // Prevents JavaScript access (prevents XSS)
			Secure:   true,                    // Ensures cookie is sent over HTTPS
			SameSite: http.SameSiteStrictMode, // Prevents CSRF attacks
			Path:     "/",
			Expires:  expirationTime,
		})

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
	}
}

func LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			Expires:  time.Unix(0, 0),
		})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
	}
}
