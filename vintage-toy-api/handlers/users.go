package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Joseph_Bartram8/vintage-toy-api/middleware"
	"github.com/Joseph_Bartram8/vintage-toy-api/models"
	"golang.org/x/crypto/bcrypt"

	//"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	//"golang.org/x/crypto/bcrypt"
)

// Get all users
func GetUsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT 
				ub.display_name, ub.store_name, ub.bio_description, ub.profile_image
			FROM users u
			JOIN user_bios ub ON u.id = ub.user_id
			WHERE u.is_deleted = FALSE
		`)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []models.PublicUserSummary
		for rows.Next() {
			var user models.PublicUserSummary
			var storeName, bioDescription, profileImage sql.NullString

			err := rows.Scan(&user.DisplayName, &storeName, &bioDescription, &profileImage)
			if err != nil {
				http.Error(w, "Error scanning users", http.StatusInternalServerError)
				return
			}

			if storeName.Valid {
				user.StoreName = &storeName.String
			}
			if bioDescription.Valid {
				user.BioDescription = &bioDescription.String
			}
			if profileImage.Valid {
				user.ProfileImage = &profileImage.String
			}

			users = append(users, user)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

// Create a new user
func CreateUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateUserRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if err := models.Validate.Struct(req); err != nil {
			http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		userID := uuid.New()

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(`
			INSERT INTO users (id, first_name, last_name, email, password_hash, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
		`, userID, req.FirstName, req.LastName, req.Email, string(hashedPassword))
		if err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "duplicate key") {
				http.Error(w, "Email already in use", http.StatusConflict)
			} else {
				http.Error(w, "Error creating user", http.StatusInternalServerError)
			}
			return
		}

		_, err = tx.Exec(`
			INSERT INTO user_bios (user_id, display_name, bio_description, profile_image, updated_at)
			VALUES ($1, $2, '', '', NOW())
		`, userID, req.DisplayName)
		if err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "duplicate key") {
				http.Error(w, "Display name already in use", http.StatusConflict)
			} else {
				http.Error(w, "Error creating user bio", http.StatusInternalServerError)
			}
			return
		}

		if err = tx.Commit(); err != nil {
			http.Error(w, "Database commit error", http.StatusInternalServerError)
			return
		}

		// Success response
		resp := map[string]interface{}{
			"id":           userID,
			"first_name":   req.FirstName,
			"last_name":    req.LastName,
			"email":        req.Email,
			"display_name": req.DisplayName,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// GetCurrentUserHandler fetches the authenticated user's info
func GetCurrentUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the JWT token from cookies
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value

		// Parse and validate the token
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Extract user ID from token claims
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			http.Error(w, "Invalid token subject", http.StatusUnauthorized)
			return
		}

		// Query user and bio data
		var user models.UserResponse
		var bio models.UserBioResponse
		var storeName, bioDescription, profileImage, updatedAt sql.NullString
		var isDeleted bool

		err = db.QueryRow(`
			SELECT u.first_name, u.last_name, u.email, u.is_deleted, 
				   ub.display_name, ub.store_name, ub.bio_description, 
				   ub.profile_image, ub.show_real_name, ub.updated_at
			FROM users u
			LEFT JOIN user_bios ub ON u.id = ub.user_id
			WHERE u.id = $1 AND u.is_deleted = FALSE;
		`, userID).Scan(
			&user.FirstName, &user.LastName, &user.Email, &isDeleted,
			&bio.DisplayName, &storeName, &bioDescription,
			&profileImage, &bio.ShowRealName, &updatedAt,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Error fetching user: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Assign nullable fields
		if storeName.Valid {
			bio.StoreName = &storeName.String
		}
		if bioDescription.Valid {
			bio.BioDescription = &bioDescription.String
		}
		if profileImage.Valid {
			bio.ProfileImage = &profileImage.String
		}
		if updatedAt.Valid {
			bio.UpdatedAt = updatedAt.String
		}

		user.IsDeleted = &isDeleted
		user.UserBio = &bio

		// Respond with JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

// UpdateUserHandler updates the authenticated user's profile
func UpdateUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from middleware context
		userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse incoming JSON
		var req models.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Start DB transaction
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Could not start transaction", http.StatusInternalServerError)
			return
		}

		// Update user_bios table
		_, err = tx.Exec(`
			UPDATE user_bios
			SET display_name = COALESCE($1, display_name),
				bio_description = COALESCE($2, bio_description),
				profile_image = COALESCE($3, profile_image),
				show_real_name = COALESCE($4, show_real_name),
				updated_at = NOW()
			WHERE user_id = $5
		`, req.DisplayName, req.BioDescription, req.ProfileImage, req.ShowRealName, userID)

		if err != nil {
			tx.Rollback()
			log.Printf("Update user_bios error: %v", err)
			http.Error(w, "Failed to update user bio", http.StatusInternalServerError)
			return
		}

		// Commit changes
		if err := tx.Commit(); err != nil {
			log.Printf("Transaction commit error: %v", err)
			http.Error(w, "Commit failed", http.StatusInternalServerError)
			return
		}

		// Return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Profile updated successfully",
		})
	}
}

// DeleteUserHandler soft-deletes the authenticated user's account
func DeleteUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from context
		userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Perform soft delete
		_, err := db.Exec("UPDATE users SET is_deleted = TRUE WHERE id = $1", userID)
		if err != nil {
			http.Error(w, "Error deleting account", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Account deleted successfully"})
	}
}

// Search a user by display name or store name
func SearchUsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Missing search query", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(`
			SELECT ub.display_name, ub.profile_image, ub.store_name
			FROM users u
			JOIN user_bios ub ON u.id = ub.user_id
			WHERE u.is_deleted = FALSE
			  AND (ub.display_name ILIKE '%' || $1 || '%' OR ub.store_name ILIKE '%' || $1 || '%')
			LIMIT 10
		`, query)
		if err != nil {
			http.Error(w, "Database query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var results []models.SearchUserResult
		for rows.Next() {
			var user models.SearchUserResult
			var storeName, profileImage sql.NullString

			if err := rows.Scan(&user.DisplayName, &profileImage, &storeName); err != nil {
				http.Error(w, "Error scanning results", http.StatusInternalServerError)
				return
			}

			if profileImage.Valid {
				user.ProfileImage = &profileImage.String
			}
			if storeName.Valid {
				user.StoreName = &storeName.String
			}

			results = append(results, user)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}
