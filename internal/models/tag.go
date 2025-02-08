package models

import "time"

type Tag struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Colour      string    `json:"colour"`
    Description string    `json:"description"`
    Items       []Item    `json:"items"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
