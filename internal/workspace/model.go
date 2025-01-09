package workspace

type CreateWorkspaceRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type UpdateWorkspaceRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    ContainerIDs []int  `json:"containerIds,omitempty"`
}