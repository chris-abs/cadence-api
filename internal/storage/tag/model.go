package tag

import "github.com/chrisabs/cadence/internal/models"

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
    TagIDs  []models.TagID  `json:"tagIds"`
    ItemIDs []models.ItemID `json:"itemIds"`
}