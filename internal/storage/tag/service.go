package tag

import (
	"fmt"
	"time"

	"github.com/chrisabs/storage/internal/storage/entities"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateTag(familyID int, req *CreateTagRequest) (*entities.Tag, error) {
    tag := &entities.Tag{
        Name:      req.Name,
        Colour:    req.Colour,
        FamilyID:  familyID,
        Items:     make([]entities.Item, 0),
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
    }

    if err := s.repo.Create(tag); err != nil {
        return nil, fmt.Errorf("failed to create tag: %v", err)
    }

    return s.repo.GetByID(tag.ID, familyID)
}

func (s *Service) GetTagByID(id int, familyID int) (*entities.Tag, error) {
    return s.repo.GetByID(id, familyID)
}

func (s *Service) GetAllTags(familyID int) ([]*entities.Tag, error) {
    return s.repo.GetByFamilyID(familyID)
}

func (s *Service) UpdateTag(id int, familyID int, req *UpdateTagRequest) (*entities.Tag, error) {
    tag, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, fmt.Errorf("tag not found: %v", err)
    }

    tag.Name = req.Name
    tag.Colour = req.Colour
    tag.UpdatedAt = time.Now().UTC()

    if err := s.repo.Update(tag); err != nil {
        return nil, fmt.Errorf("failed to update tag: %v", err)
    }

    return s.repo.GetByID(id, familyID)
}

func (s *Service) AssignTagsToItems(familyID int, tagIDs []int, itemIDs []int) error {
    return s.repo.AssignTagsToItems(familyID, tagIDs, itemIDs)
}

func (s *Service) DeleteTag(id int, familyID int) error {
    if _, err := s.repo.GetByID(id, familyID); err != nil {
        return fmt.Errorf("tag not found: %v", err)
    }
    return s.repo.Delete(id, familyID)
}