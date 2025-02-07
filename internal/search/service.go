package search

import (
	"fmt"

	"github.com/chrisabs/storage/internal/models"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{
        repo: repo,
    }
}

func (s *Service) Search(query string, userID int) (*SearchResponse, error) {
    if query == "" {
        return nil, fmt.Errorf("search query cannot be empty")
    }

    results, err := s.repo.Search(query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to execute search: %v", err)
    }

    return results, nil
}

func (s *Service) SearchWorkspaces(query string, userID int) (WorkspaceSearchResults, error) {
    if query == "" {
        return nil, fmt.Errorf("search query cannot be empty")
    }

    results, err := s.repo.SearchWorkspaces(query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to execute workspace search: %v", err)
    }

    return results, nil
}

func (s *Service) SearchContainers(query string, userID int) (ContainerSearchResults, error) {
    if query == "" {
        return nil, fmt.Errorf("search query cannot be empty")
    }

    results, err := s.repo.SearchContainers(query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to execute container search: %v", err)
    }

    return results, nil
}

func (s *Service) SearchItems(query string, userID int) (ItemSearchResults, error) {
    if query == "" {
        return nil, fmt.Errorf("search query cannot be empty")
    }

    results, err := s.repo.SearchItems(query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to execute item search: %v", err)
    }

    return results, nil
}

func (s *Service) SearchTags(query string, userID int) (TagSearchResults, error) {
    if query == "" {
        return nil, fmt.Errorf("search query cannot be empty")
    }

    results, err := s.repo.SearchTags(query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to execute tag search: %v", err)
    }

    return results, nil
}

func (s *Service) FindContainerByQR(qrCode string, userID int) (*models.Container, error) {
    if qrCode == "" {
        return nil, fmt.Errorf("QR code cannot be empty")
    }

    container, err := s.repo.FindContainerByQR(qrCode, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to find container: %v", err)
    }

    return container, nil
}