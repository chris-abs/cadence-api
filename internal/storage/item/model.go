package item

import "github.com/chrisabs/cadence/internal/models"

type CreateItemRequest struct {
    Name        string              `json:"name"`
    Description string              `json:"description"`
    Quantity    int                 `json:"quantity"`
    ContainerID *models.ContainerID `json:"containerId,omitempty"`
    TagNames    []string            `json:"tagNames"`
}

type UpdateItemRequest struct {
    Name           string              `json:"name"`
    Description    string              `json:"description"`
    Quantity       int                 `json:"quantity"`
    ContainerID    *models.ContainerID `json:"containerId,omitempty"`
    Tags           []models.TagID      `json:"tags,omitempty"`
    ImagesToDelete []string            `json:"imagesToDelete,omitempty"`
}

type AddImageRequest struct {
    ItemID   models.ItemID `json:"itemId"`
    ImageURL string        `json:"imageUrl"`
}