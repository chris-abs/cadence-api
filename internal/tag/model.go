package tag

type CreateTagRequest struct {
	Name   string `json:"name"`
	Colour string `json:"colour"`
}

type UpdateTagRequest struct {
	Name   string `json:"name"`
	Colour string `json:"colour"`
}

type BulkAssignRequest struct {
    TagIDs  []int `json:"tagIds"`
    ItemIDs []int `json:"itemIds"`
}