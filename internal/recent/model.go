package recent

import "time"

type EntityPreview struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

type EntityStats struct {
    Recent []EntityPreview `json:"recent"`
    Total  int            `json:"total"`
}

type Response struct {
    Workspaces EntityStats `json:"workspaces"`
    Containers EntityStats `json:"containers"`
    Items      EntityStats `json:"items"`
    Tags       EntityStats `json:"tags"`
}