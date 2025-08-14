package workspace

import "github.com/chrisabs/cadence/internal/models"

type CreateWorkspaceRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type UpdateWorkspaceRequest struct {
    Name         string                `json:"name"`
    Description  string                `json:"description"`
    ContainerIDs []models.ContainerID  `json:"containerIds,omitempty"`
}