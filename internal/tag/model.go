package tag

type CreateTagRequest struct {
	Name   string `json:"name"`
	Colour string `json:"colour"`
}

type UpdateTagRequest struct {
	Name   string `json:"name"`
	Colour string `json:"colour"`
}
