package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"fmt"

	"github.com/Joseph_Bartram8/vintage-toy-api/middleware"
	"github.com/Joseph_Bartram8/vintage-toy-api/models"

	//"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Get all users
func GetUsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query users with privacy setting applied
		rows, err := db.Query(`
			SELECT 
				u.is_deleted, ub.display_name, ub.store_name, ub.bio_description, 
				ub.profile_image, ub.show_real_name, ub.updated_at,
				CASE WHEN ub.show_real_name THEN u.first_name ELSE NULL END AS first_name,
				CASE WHEN ub.show_real_name THEN u.last_name ELSE NULL END AS last_name
			FROM users u
			JOIN user_bios ub ON u.id = ub.user_id
			WHERE u.is_deleted = FALSE;
		`)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Parse results into a slice
		var users []models.User
		for rows.Next() {
			var user models.User
			var bio models.UserBioResponse
			var firstName, lastName, storeName, bioDescription, profileImage, updatedAt sql.NullString

			err := rows.Scan(
				&user.IsDeleted,
				&bio.DisplayName, &storeName, &bioDescription,
				&bio.ProfileImage, &bio.ShowRealName, &updatedAt,
				&firstName, &lastName,
			)
			if err != nil {
				http.Error(w, "Error scanning users", http.StatusInternalServerError)
				return
			}

			// Assign nullable fields safely
			if firstName.Valid {
				user.FirstName = firstName.String
			}
			if lastName.Valid {
				user.LastName = lastName.String
			}
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

			// Attach UserBioResponse to UserResponse
			user.UserBio = &bio

			users = append(users, user)
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

// Create a new user
func CreateUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateUserRequest

		// Decode JSON request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Validate user input
		if err := models.Validate.Struct(req); err != nil {
			http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// Generate a UUID for the new user
		userID := uuid.New()

		// Insert user into the database
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Insert into users table
		_, err = tx.Exec(`
			INSERT INTO users (id, first_name, last_name, email, password_hash, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())`,
			userID, req.FirstName, req.LastName, req.Email, string(hashedPassword),
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}

		// Insert into user_bios table
		_, err = tx.Exec(`
			INSERT INTO user_bios (user_id, display_name, bio_description, profile_image, updated_at)
			VALUES ($1, $2, '', '', NOW())`,
			userID, req.DisplayName,
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Error creating user bio", http.StatusInternalServerError)
			return
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Database commit error", http.StatusInternalServerError)
			return
		}

		// Return created user info (excluding password)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           userID,
			"first_name":   req.FirstName,
			"last_name":    req.LastName,
			"email":        req.Email,
			"display_name": req.DisplayName,
		})
	}
}

/*func GetUserByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from URL
		userIDStr := chi.URLParam(r, "id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Fetch user data
		var user models.User
		err = db.QueryRow("SELECT id, first_name, last_name, email, created_at FROM users WHERE id=$1", userID).
			Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(user)
	}
}*/

// GetCurrentUserHandler fetches the authenticated user's info
func GetCurrentUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received Cookies:", r.Cookies())

		// Retrieve the JWT token from cookies
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			log.Println("No auth_token cookie found")
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value
		fmt.Println("Cookie Value:", tokenStr)

		// Parse and validate the token
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			log.Println("Invalid or expired token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract user ID from token
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			log.Println("Invalid token subject")
			http.Error(w, "Invalid token subject", http.StatusUnauthorized)
			return
		}

		// Query user details along with user bio in one query
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
			log.Println("User not found in database")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("Database error:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Assign nullable fields safely
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

		// Attach UserBioResponse to UserResponse
		user.IsDeleted = &isDeleted
		user.UserBio = &bio

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

// UpdateUserHandler updates the authenticated user's profile
func UpdateUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from context (already correct)
		userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse request body
		var req models.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Start a database transaction
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Database transaction error", http.StatusInternalServerError)
			return
		}

		// Update `users` table if needed
		if req.FirstName != nil || req.LastName != nil {
			_, err := tx.Exec(`
				UPDATE users 
				SET first_name = COALESCE($1, first_name), 
					last_name = COALESCE($2, last_name)
				WHERE id = $3`,
				req.FirstName, req.LastName, userID,
			)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Error updating user profile", http.StatusInternalServerError)
				return
			}
		}

		// Update `user_bios` table if needed
		if req.DisplayName != nil || req.BioDescription != nil || req.ProfileImage != nil {
			_, err := tx.Exec(`
				UPDATE user_bios 
				SET display_name = COALESCE($1, display_name), 
					bio_description = COALESCE($2, bio_description), 
					profile_image = COALESCE($3, profile_image),
					updated_at = NOW()
				WHERE user_id = $4`,
				req.DisplayName, req.BioDescription, req.ProfileImage, userID,
			)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Error updating user bio", http.StatusInternalServerError)
				return
			}
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Database commit error", http.StatusInternalServerError)
			return
		}

		// Return success message
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
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
