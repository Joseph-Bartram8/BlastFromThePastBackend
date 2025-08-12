package models

import (
	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

// public list of users
type PublicUserSummary struct {
	DisplayName    string  `json:"display_name"`
	StoreName      *string `json:"store_name,omitempty"`
	BioDescription *string `json:"bio_description,omitempty"`
	ProfileImage   *string `json:"profile_image,omitempty"`
}

// CreateUserRequest struct
type CreateUserRequest struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
}

// UserResponse struct
type UserResponse struct {
	ID        *string          `json:"uuid"`
	FirstName *string          `json:"first_name"`
	LastName  *string          `json:"last_name"`
	Email     *string          `json:"email"`
	IsDeleted *bool            `json:"is_deleted"`
	UserBio   *UserBioResponse `json:"user_bio,omitempty"`
}

// UserBioResponse struct
type UserBioResponse struct {
	DisplayName    string  `json:"display_name" validate:"required"`
	StoreName      *string `json:"store_name,omitempty"`
	BioDescription *string `json:"bio_description,omitempty"`
	ProfileImage   *string `json:"profile_image,omitempty"`
	ShowRealName   *bool   `json:"show_real_name,omitempty"`
	UpdatedAt      string  `json:"updated_at"`
}

// SearchUserRequest struct
type SearchUserResult struct {
	DisplayName  string  `json:"display_name"`
	ProfileImage *string `json:"profile_image,omitempty"`
	StoreName    *string `json:"store_name,omitempty"`
}
