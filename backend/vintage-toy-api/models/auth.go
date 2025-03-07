package models

// LoginRequest struct
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse struct
type LoginResponse struct {
	Token string `json:"token"`
}

// UpdateUserRequest struct
type UpdateUserRequest struct {
	FirstName      *string `json:"first_name,omitempty"`
	LastName       *string `json:"last_name,omitempty"`
	DisplayName    *string `json:"display_name,omitempty"`
	BioDescription *string `json:"bio_description,omitempty"`
	ProfileImage   *string `json:"profile_image,omitempty"`
}
