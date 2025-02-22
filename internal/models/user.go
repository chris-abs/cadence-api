package models

import "time"

type UserRole string

const (
    RoleParent UserRole = "PARENT"
    RoleChild  UserRole = "CHILD"
)

type User struct {
    ID         int         `json:"id"`
    Email      string      `json:"email"`
    Password   string      `json:"-"`
    FirstName  string      `json:"firstName"`
    LastName   string      `json:"lastName"`
    ImageURL   string      `json:"imageUrl"`
    FamilyID   *int        `json:"familyId"`
    Role       *UserRole   `json:"role"`
    Containers []Container `json:"containers"`
    CreatedAt  time.Time   `json:"createdAt"`
    UpdatedAt  time.Time   `json:"updatedAt"`
}