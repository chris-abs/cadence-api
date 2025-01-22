package container

type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	Quantity    int    `json:"quantity"`
	TagIDs      []int  `json:"tagIds"`
}

type CreateContainerRequest struct {
    Name        string              `json:"name"`
    Location    string              `json:"location"`
    WorkspaceID *int                 `json:"workspaceId,omitempty"`
    Items       []CreateItemRequest `json:"items"`
}

type UpdateContainerRequest struct {
    Name        string `json:"name"`
    Location    string `json:"location"`
    WorkspaceID *int    `json:"workspaceId,omitempty"`
    ItemIDs     []int  `json:"itemIds,omitempty"`
}
