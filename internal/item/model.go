package item

type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	Quantity    int    `json:"quantity"`
	ContainerID *int   `json:"containerId,omitempty"`
	TagIDs      []int  `json:"tagIds"`
}
