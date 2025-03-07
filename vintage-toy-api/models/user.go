package models

import (
	"github.com/go-playground/validator/v10"
)

// Public User struct
type User struct {
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	IsDeleted bool             `json:"is_deleted"`
	UserBio   *UserBioResponse `json:"user_bio,omitempty"`
}

var Validate = validator.New()

// CreateUserRequest struct
type CreateUserRequest struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required"`
}

// UserResponse struct
type UserResponse struct {
	FirstName *string          `json:"first_name" validate:"required"`
	LastName  *string          `json:"last_name" validate:"required"`
	Email     *string          `json:"email" validate:"required,email"`
	IsDeleted *bool            `json:"is_deleted"`
	UserBio   *UserBioResponse `json:"user_bio,omitempty"`
}

// UserBioResponse struct
type UserBioResponse struct {
	DisplayName    string  `json:"display_name" validate:"required"`
	StoreName      *string `json:"store_name,omitempty"`
	BioDescription *string `json:"bio_description,omitempty"`
	ProfileImage   *string `json:"profile_image,omitempty"`
	ShowRealName   bool    `json:"show_real_name"`
	UpdatedAt      string  `json:"updated_at"`
}
