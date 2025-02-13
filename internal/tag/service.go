package tag

import (
	"fmt"
	"time"

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

func (s *Service) CreateTag(req *CreateTagRequest) (*models.Tag, error) {
	tag := &models.Tag{
		Name:      req.Name,
		Colour:    req.Colour,
		Items:     make([]models.Item, 0),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(tag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %v", err)
	}

	return s.repo.GetByID(tag.ID)
}

func (s *Service) GetTagByID(id int) (*models.Tag, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetAllTags() ([]*models.Tag, error) {
	return s.repo.GetAll()
}

func (s *Service) UpdateTag(id int, req *UpdateTagRequest) (*models.Tag, error) {
	tag, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("tag not found: %v", err)
	}

	tag.Name = req.Name
	tag.Colour = req.Colour
	tag.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(tag); err != nil {
		return nil, fmt.Errorf("failed to update tag: %v", err)
	}

	return s.repo.GetByID(id)
}

func (s *Service) BulkAssignTags(tagIDs []int, itemIDs []int) error {
    return s.repo.BulkAssignTags(tagIDs, itemIDs)
}

func (s *Service) DeleteTag(id int) error {
	return s.repo.Delete(id)
}
