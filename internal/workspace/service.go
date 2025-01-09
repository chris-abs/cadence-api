package workspace

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chrisabs/storage/internal/models"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateWorkspace(userID int, req *CreateWorkspaceRequest) (*models.Workspace, error) {
    workspace := &models.Workspace{
        ID:          rand.Intn(10000),
        Name:        req.Name,
        Description: req.Description,
        UserID:      userID,
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
        Containers:  make([]models.Container, 0),
    }

    if err := s.repo.Create(workspace); err != nil {
        return nil, fmt.Errorf("failed to create workspace: %v", err)
    }

    return s.repo.GetByID(workspace.ID)
}

func (s *Service) GetWorkspaceByID(id int) (*models.Workspace, error) {
    workspace, err := s.repo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("error getting workspace: %v", err)
    }
    return workspace, nil
}

func (s *Service) GetWorkspacesByUserID(userID int) ([]*models.Workspace, error) {
    return s.repo.GetByUserID(userID)
}

func (s *Service) UpdateWorkspace(id int, req *UpdateWorkspaceRequest) (*models.Workspace, error) {
    workspace, err := s.repo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("workspace not found: %v", err)
    }

    workspace.Name = req.Name
    workspace.Description = req.Description
    workspace.UpdatedAt = time.Now().UTC()

    if err := s.repo.Update(workspace); err != nil {
        return nil, fmt.Errorf("failed to update workspace: %v", err)
    }

    return workspace, nil
}

func (s *Service) DeleteWorkspace(id int) error {
    return s.repo.Delete(id)
}