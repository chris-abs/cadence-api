package workspace

import (
	"fmt"
	"math/rand"
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

func (s *Service) CreateWorkspace(profileCtx *models.ProfileContext, req *CreateWorkspaceRequest) (*entities.Workspace, error) {
    workspace := &entities.Workspace{
        ID:          rand.Intn(10000),
        Name:        req.Name,
        Description: req.Description,
        profileId:      profileCtx.ProfileID,
        FamilyID:    profileCtx.FamilyID,
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
        Containers:  make([]entities.Container, 0),
    }

    if err := s.repo.Create(workspace); err != nil {
        return nil, fmt.Errorf("failed to create workspace: %v", err)
    }

    return s.repo.GetByID(workspace.ID, profileCtx.FamilyID)
}

func (s *Service) GetWorkspaceByID(id int, familyID int) (*entities.Workspace, error) {
    workspace, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, fmt.Errorf("error getting workspace: %v", err)
    }
    return workspace, nil
}

func (s *Service) GetWorkspacesByFamilyID(familyID int, profileId int) ([]*entities.Workspace, error) {
    return s.repo.GetByFamilyID(familyID, profileId)
}

func (s *Service) UpdateWorkspace(id int, familyID int, req *UpdateWorkspaceRequest) (*entities.Workspace, error) {
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

func (s *Service) DeleteWorkspace(id int, familyID int, deletedBy int) error {
    if err := s.repo.Delete(id, familyID, deletedBy); err != nil {
        return fmt.Errorf("failed to delete workspace: %v", err)
    }
    return nil
}

func (s *Service) RestoreWorkspace(id int, familyID int) error {
    if err := s.repo.RestoreDeleted(id, familyID); err != nil {
        return fmt.Errorf("failed to restore workspace: %v", err)
    }
    return nil
}