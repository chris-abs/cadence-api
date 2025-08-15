package sources

import (
	"fmt"

	"github.com/chrisabs/cadence/internal/media/sources/entities"
	"github.com/chrisabs/cadence/internal/models"
	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSource(req *entities.CreateSourceRequest) (*entities.Source, error) {
	if err := s.validateCreateSourceRequest(req); err != nil {
		return nil, err
	}
	
	source := &entities.Source{
		ID:       models.SourceID(uuid.New().String()),
		Name:     req.Name,
		Color:    req.Color,
		Category: req.Category,
	}
	
	if err := s.repo.CreateSource(source); err != nil {
		return nil, fmt.Errorf("error creating source: %v", err)
	}
	
	return source, nil
}

func (s *Service) GetSourceByID(sourceID models.SourceID) (*entities.Source, error) {
	return s.repo.GetSourceByID(sourceID)
}

func (s *Service) UpdateSource(sourceID models.SourceID, req *entities.UpdateSourceRequest) (*entities.Source, error) {
	if err := s.validateUpdateSourceRequest(req); err != nil {
		return nil, err
	}
	
	existingSource, err := s.repo.GetSourceByID(sourceID)
	if err != nil {
		return nil, err
	}
	
	if req.Name != nil {
		existingSource.Name = *req.Name
	}
	if req.Color != nil {
		existingSource.Color = *req.Color
	}
	if req.Category != nil {
		existingSource.Category = *req.Category
	}
	
	if err := s.repo.UpdateSource(existingSource); err != nil {
		return nil, fmt.Errorf("error updating source: %v", err)
	}
	
	return existingSource, nil
}

func (s *Service) DeleteSource(sourceID models.SourceID, deletedBy models.ProfileID) error {
	count, err := s.repo.GetMediaCountBySource(sourceID)
	if err != nil {
		return fmt.Errorf("error checking source usage: %v", err)
	}
	
	if count > 0 {
		return fmt.Errorf("cannot delete source: it is being used by %d media items", count)
	}
	
	return s.repo.DeleteSource(sourceID, deletedBy)
}

func (s *Service) GetAllSources(params entities.SourceSearchParams) (*entities.SourceSearchResponse, error) {
	return s.repo.GetAllSources(params)
}

func (s *Service) validateCreateSourceRequest(req *entities.CreateSourceRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	
	if len(req.Name) > 100 {
		return fmt.Errorf("name must be 100 characters or less")
	}
	
	if req.Color == "" {
		return fmt.Errorf("color is required")
	}
	
	if req.Category == "" {
		return fmt.Errorf("category is required")
	}
	
	if len(req.Category) > 50 {
		return fmt.Errorf("category must be 50 characters or less")
	}
	
	return nil
}

func (s *Service) validateUpdateSourceRequest(req *entities.UpdateSourceRequest) error {
	if req.Name != nil && len(*req.Name) > 100 {
		return fmt.Errorf("name must be 100 characters or less")
	}
	
	if req.Category != nil && len(*req.Category) > 50 {
		return fmt.Errorf("category must be 50 characters or less")
	}
	
	return nil
}