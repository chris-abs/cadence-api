package tag

type CreateTagRequest struct {
	Name   		string `json:"name"`
	Description string `json:"description"`
	Colour 		string `json:"colour"`
}

type UpdateTagRequest struct {
	Name        string `json:"name"`
	Colour      string `json:"colour"`
	Description string `json:"description"`
}

type AssignTagsRequest struct {
    TagIDs  []int `json:"tagIds"`
    ItemIDs []int `json:"itemIds"`
}