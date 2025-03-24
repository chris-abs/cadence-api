package entities

import "time"

type ItemImage struct {
    ID           int       `json:"id"`
    URL          string    `json:"url"`
    DisplayOrder int       `json:"displayOrder"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
}

type Item struct {
    ID          int         `json:"id"`
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Images      []ItemImage `json:"images"`
    Quantity    int         `json:"quantity"`
    ContainerID *int        `json:"containerId,omitempty"`
    Container   *Container  `json:"container,omitempty"`
    ProfileID   int         `json:"profileId"`
    FamilyID    int         `json:"familyId"`    
    Tags        []Tag       `json:"tags"`
    CreatedAt   time.Time   `json:"createdAt"`
    UpdatedAt   time.Time   `json:"updatedAt"`
}
