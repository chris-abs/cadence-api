package container

type CreateContainerRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	ItemIDs  []int  `json:"itemIds,omitempty"`
}

type UpdateContainerRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	ItemIDs  []int  `json:"itemIds,omitempty"`
}
