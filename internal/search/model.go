package search

type SearchResult struct {
	Type        string  `json:"type"`
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Rank        float64 `json:"rank"`
}

type SearchResponse struct {
	Containers []SearchResult `json:"containers"`
	Items      []SearchResult `json:"items"`
	Tags       []SearchResult `json:"tags"`
}
