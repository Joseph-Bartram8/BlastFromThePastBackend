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
	DisplayName    *string `json:"display_name,omitempty"`
	BioDescription *string `json:"bio_description,omitempty"`
	ProfileImage   *string `json:"profile_image,omitempty"`
	ShowRealName   *bool   `json:"show_real_name,omitempty"`
}
