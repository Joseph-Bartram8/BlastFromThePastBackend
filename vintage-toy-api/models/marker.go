package models

import "time"

// MarkerResponse represents the structure of a marker returned by the API
type MarkerResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Latitude    float64         `json:"latitude"`
	Longitude   float64         `json:"longitude"`
	Region      string          `json:"region"`
	MarkerType  string          `json:"marker_type"`
	CreatedAt   time.Time       `json:"created_at"`
	User        MarkerUserInfo  `json:"user"`
}

// MarkerUserInfo holds the user details associated with the marker
type MarkerUserInfo struct {
	DisplayName string  `json:"display_name"`
	StoreName   *string `json:"store_name,omitempty"`
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	ProfileImage *string `json:"profile_image,omitempty"`
}
