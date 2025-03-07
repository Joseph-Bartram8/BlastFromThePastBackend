package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Joseph_Bartram8/vintage-toy-api/models"
)

// GetAllMarkersHandler retrieves all markers along with relevant user data, including profile image
func GetAllMarkersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
			SELECT 
				um.id, um.name, um.description, um.latitude, um.longitude, um.region, um.marker_type, um.created_at,
				ub.display_name, ub.store_name, u.first_name, u.last_name, ub.show_real_name, ub.profile_image
			FROM user_markers um
			JOIN users u ON um.user_id = u.id
			LEFT JOIN user_bios ub ON u.id = ub.user_id
			WHERE u.is_deleted = FALSE;
		`

		rows, err := db.Query(query)
		if err != nil {
			log.Println("Database query error:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var markers []models.MarkerResponse

		for rows.Next() {
			var marker models.MarkerResponse
			var user models.MarkerUserInfo
			var firstName, lastName, profileImage sql.NullString
			var showRealName bool

			err := rows.Scan(
				&marker.ID, &marker.Name, &marker.Description, &marker.Latitude, &marker.Longitude,
				&marker.Region, &marker.MarkerType, &marker.CreatedAt,
				&user.DisplayName, &user.StoreName, &firstName, &lastName, &showRealName, &profileImage,
			)

			if err != nil {
				log.Println("Row scan error:", err)
				http.Error(w, "Database scan error", http.StatusInternalServerError)
				return
			}

			// Handle privacy setting for real names
			if showRealName {
				if firstName.Valid {
					user.FirstName = &firstName.String
				}
				if lastName.Valid {
					user.LastName = &lastName.String
				}
			}

			// Assign profile image if available
			if profileImage.Valid {
				user.ProfileImage = &profileImage.String
			}

			marker.User = user
			markers = append(markers, marker)
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(markers)
	}
}
