package models

import "time"

type User struct {
    ID         int         `json:"id"`
    Email      string      `json:"email"`
    Password   string      `json:"-"`
    FirstName  string      `json:"firstName"`
    LastName   string      `json:"lastName"`
    ImageURL   string      `json:"imageUrl"`
    Role       UserRole    `json:"role"`
    FamilyID   *int        `json:"familyId"`
    Containers []Container `json:"containers"`
    CreatedAt  time.Time   `json:"createdAt"`
    UpdatedAt  time.Time   `json:"updatedAt"`
}

type UserRole string

const (
    RoleParent UserRole = "PARENT"
    RoleChild  UserRole = "CHILD"
)
