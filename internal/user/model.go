package user

import "github.com/chrisabs/storage/internal/models"

type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	ImageURL  string `json:"imageUrl"`
}

type UpdateUserRequest struct {
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
    ImageURL  string `json:"imageUrl,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}
