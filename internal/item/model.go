package item

type CreateItemRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ImageURL    string   `json:"imageUrl"`
	Quantity    int      `json:"quantity"`
	ContainerID *int     `json:"containerId,omitempty"`
	TagNames    []string `json:"tagNames"`
}
type UpdateItemRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    ImageURL    string `json:"imageUrl"`
    Quantity    int    `json:"quantity"`
    ContainerID *int   `json:"containerId,omitempty"`
    Tags        []int  `json:"tags,omitempty"`
}