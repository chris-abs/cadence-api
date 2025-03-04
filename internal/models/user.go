package models

import (
	"time"

	"github.com/chrisabs/storage/internal/storage/entities"
)

type UserRole string

const (
    RoleParent UserRole = "PARENT"
    RoleChild  UserRole = "CHILD"
)

type User struct {
    ID         int                  `json:"id"`
    Email      string               `json:"email"`
    Password   string               `json:"-"`
    FirstName  string               `json:"firstName"`
    LastName   string               `json:"lastName"`
    ImageURL   string               `json:"imageUrl"`
    Containers []entities.Container `json:"containers,omitempty"`
    CreatedAt  time.Time            `json:"createdAt"`
    UpdatedAt  time.Time            `json:"updatedAt"`
}