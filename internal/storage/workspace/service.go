package workspace

import (
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/models"
	"github.com/chrisabs/cadence/internal/storage/entities"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateWorkspace(familyID models.FamilyID, profileID models.ProfileID, req *CreateWorkspaceRequest) (*entities.Workspace, error) {
    workspace := &entities.Workspace{
        Name:        req.Name,
        Description: req.Description,
        ProfileID:   profileID,
        FamilyID:    familyID,
        Containers:  make([]entities.Container, 0),
    }

    if err := s.repo.Create(workspace); err != nil {
        return nil, fmt.Errorf("failed to create workspace: %v", err)
    }

    return s.repo.GetByID(workspace.ID, familyID)
}

func (s *Service) GetWorkspaceByID(id models.WorkspaceID, familyID models.FamilyID) (*entities.Workspace, error) {
    workspace, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, fmt.Errorf("error getting workspace: %v", err)
    }
    return workspace, nil
}

func (s *Service) GetWorkspacesByFamilyID(familyID models.FamilyID, profileID models.ProfileID) ([]*entities.Workspace, error) {
    return s.repo.GetByFamilyID(familyID, profileID)
}

func (s *Service) UpdateWorkspace(id models.WorkspaceID, familyID models.FamilyID, req *UpdateWorkspaceRequest) (*entities.Workspace, error) {
    workspace, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, fmt.Errorf("workspace not found: %v", err)
    }

    workspace.Name = req.Name
    workspace.Description = req.Description
    workspace.UpdatedAt = time.Now().UTC()

    if err := s.repo.Update(workspace); err != nil {
        return nil, fmt.Errorf("failed to update workspace: %v", err)
    }

    if len(req.ContainerIDs) > 0 {
        if err := s.repo.UpdateContainers(workspace.ID, familyID, req.ContainerIDs); err != nil {
            return nil, fmt.Errorf("failed to update container assignments: %v", err)
        }
    }

    return s.repo.GetByID(workspace.ID, familyID)
}

func (s *Service) DeleteWorkspace(id models.WorkspaceID, familyID models.FamilyID, deletedBy models.ProfileID) error {
    if err := s.repo.Delete(id, familyID, deletedBy); err != nil {
        return fmt.Errorf("failed to delete workspace: %v", err)
    }
    return nil
}

func (s *Service) RestoreWorkspace(id models.WorkspaceID, familyID models.FamilyID) error {
    if err := s.repo.RestoreDeleted(id, familyID); err != nil {
        return fmt.Errorf("failed to restore workspace: %v", err)
    }
    return nil
}