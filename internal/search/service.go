package search

import "fmt"

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
