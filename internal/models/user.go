package models

import "time"

type User struct {
	ID         int         `json:"id"`
	Email      string      `json:"email"`
	Password   string      `json:"-"`
	FirstName  string      `json:"firstName"`
	LastName   string      `json:"lastName"`
	ImageURL   string      `json:"imageUrl"`
	Containers []Container `json:"containers"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}
