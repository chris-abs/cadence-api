package search

import (
	"fmt"

	"github.com/chrisabs/cadence/internal/storage/entities"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{
        repo: repo,
    }
}

func (s *Service) Search(query string, familyID int) (*SearchResponse, error) {
    if query == "" {
        return nil, fmt.Errorf("search query cannot be empty")
    }

    results, err := s.repo.Search(query, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to execute search: %v", err)
    }

    return results, nil
}

func (s *Service) SearchWorkspaces(query string, familyID int) (WorkspaceSearchResults, error) {
    if query == "" {
        return nil, fmt.Errorf("search query cannot be empty")
    }

    results, err := s.repo.SearchWorkspaces(query, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to execute workspace search: %v", err)
    }

    return results, nil
}

func (s *Service) SearchContainers(query string, familyID int) (ContainerSearchResults, error) {
    if query == "" {
        return nil, fmt.Errorf("search query cannot be empty")
    }

    results, err := s.repo.SearchContainers(query, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to execute container search: %v", err)
    }

    return results, nil
}

func (s *Service) SearchItems(query string, familyID int) (ItemSearchResults, error) {
    if query == "" {
        return nil, fmt.Errorf("search query cannot be empty")
    }

    results, err := s.repo.SearchItems(query, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to execute item search: %v", err)
    }

    return results, nil
}

func (s *Service) SearchTags(query string, familyID int) (TagSearchResults, error) {
    if query == "" {
        return nil, fmt.Errorf("search query cannot be empty")
    }

    results, err := s.repo.SearchTags(query, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to execute tag search: %v", err)
    }

    return results, nil
}

func (s *Service) FindContainerByQR(qrCode string, familyID int) (*entities.Container, error) {
    if qrCode == "" {
        return nil, fmt.Errorf("QR code cannot be empty")
    }

    container, err := s.repo.FindContainerByQR(qrCode, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to find container: %v", err)
    }

    return container, nil
}