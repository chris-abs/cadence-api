package search

type SearchResult struct {
    Type          string  `json:"type"`
    ID            int     `json:"id"`
    Name          string  `json:"name"`
    Description   string  `json:"description"`
    Rank          float64 `json:"rank"`
    ContainerName *string `json:"containerName,omitempty"`
    WorkspaceName *string `json:"workspaceName,omitempty"`
}

type SearchResponse struct {
    Workspaces []SearchResult `json:"workspaces"`
    Containers []SearchResult `json:"containers"`
    Items      []SearchResult `json:"items"`
    Tags       []SearchResult `json:"tags"`
    TaggedItems []SearchResult `json:"taggedItems"`
}

type WorkspaceSearchResults []SearchResult
type ContainerSearchResults []SearchResult
type ItemSearchResults []SearchResult
type TagSearchResults []SearchResult
type TaggedItemsSearchResults []SearchResult