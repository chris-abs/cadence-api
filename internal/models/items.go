package models

import "time"

type Item struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageURL    string    `json:"imageUrl"`
	Quantity    int       `json:"quantity"`
	ContainerID *int      `json:"containerId,omitempty"`
	Container   *Container `json:"container"` 
	Tags        []Tag     `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
