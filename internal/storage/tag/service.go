package tag

import (
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/storage/entities"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateTag(familyID int, profileID int, req *CreateTagRequest) (*entities.Tag, error) {
    tag := &entities.Tag{
        Name:        req.Name,
        Description: req.Description,
        Colour:      req.Colour,
        ProfileID:   profileID,
        FamilyID:    familyID,
        Items:       make([]entities.Item, 0),
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
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

func (s *Service) UpdateTag(id int, familyID int, profileID int, req *UpdateTagRequest) (*entities.Tag, error) {
    tag, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, fmt.Errorf("tag not found: %v", err)
    }

    tag.Name = req.Name
    tag.Description = req.Description
    tag.Colour = req.Colour
    tag.ProfileID = profileID 
    tag.UpdatedAt = time.Now().UTC()

    if err := s.repo.Update(tag); err != nil {
        return nil, fmt.Errorf("failed to update tag: %v", err)
    }

    return s.repo.GetByID(id, familyID)
}

func (s *Service) AssignTagsToItems(familyID int, tagIDs []int, itemIDs []int) error {
    return s.repo.AssignTagsToItems(familyID, tagIDs, itemIDs)
}

func (s *Service) DeleteTag(id int, familyID int, deletedBy int) error {
    return s.repo.Delete(id, familyID, deletedBy)
}

func (s *Service) RestoreTag(id int, familyID int) error {
    if err := s.repo.RestoreDeleted(id, familyID); err != nil {
        return fmt.Errorf("failed to restore tag: %v", err)
    }
    return nil
}