package container

import "github.com/chrisabs/cadence/internal/models"

type CreateItemRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	ImageURL    string           `json:"imageUrl"`
	Quantity    int              `json:"quantity"`
	TagIDs      []models.TagID   `json:"tagIds"`
}

type CreateContainerRequest struct {
    Name        string                  `json:"name"`
	Description string                  `json:"description"`
    Location    string                  `json:"location"`
    WorkspaceID *models.WorkspaceID     `json:"workspaceId,omitempty"`
    Items       []CreateItemRequest     `json:"items"`
}

type UpdateContainerRequest struct {
    Name        string                  `json:"name"`
	Description string                  `json:"description"`
    Location    string                  `json:"location"`
    WorkspaceID *models.WorkspaceID     `json:"workspaceId,omitempty"`
    ItemIDs     []models.ItemID         `json:"itemIds,omitempty"`
}