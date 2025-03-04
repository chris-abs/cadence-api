package search

import (
	"github.com/chrisabs/storage/internal/storage/entities"
)

type SearchResult struct {
    Type          string  `json:"type"`
    ID            int     `json:"id"`
    Name          string  `json:"name"`
    Description   string  `json:"description"`
    Rank          float64 `json:"rank"`
    ContainerName *string `json:"containerName,omitempty"`
    WorkspaceName *string `json:"workspaceName,omitempty"`
    Colour        *string `json:"colour,omitempty"` 
}

type SearchResponse struct {
    Workspaces  []SearchResult `json:"workspaces"`
    Containers  []SearchResult `json:"containers"`
    Items       []SearchResult `json:"items"`
    Tags        []SearchResult `json:"tags"`
    TaggedItems []SearchResult `json:"taggedItems"`
}

type WorkspaceSearchResult struct {
    entities.Workspace
    Rank float64 `json:"rank"`
}

type ContainerSearchResult struct {
    entities.Container
    Rank float64 `json:"rank"`
}

type ItemSearchResult struct {
    entities.Item
    Rank float64 `json:"rank"`
}

type TagSearchResult struct {
    entities.Tag
    Rank float64 `json:"rank"`
}

type WorkspaceSearchResults []WorkspaceSearchResult
type ContainerSearchResults []ContainerSearchResult
type ItemSearchResults []ItemSearchResult
type TagSearchResults []TagSearchResult