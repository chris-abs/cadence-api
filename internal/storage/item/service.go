package item

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

func (s *Service) CreateItem(familyID int, req *CreateItemRequest) (*entities.Item, error) {
    if req.Name == "" {
        return nil, fmt.Errorf("item name is required")
    }

    item := &entities.Item{
        Name:        req.Name,
        Description: req.Description,
        Quantity:    req.Quantity,
        ContainerID: req.ContainerID,
        FamilyID:    familyID,
        Images:      []entities.ItemImage{},
        Tags:        make([]entities.Tag, 0),
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
    }

    createdItem, err := s.repo.Create(item, req.TagNames)
    if err != nil {
        return nil, fmt.Errorf("failed to create item: %v", err)
    }

    return createdItem, nil
}

func (s *Service) GetItemByID(id int, familyID int) (*entities.Item, error) {
    return s.repo.GetByID(id, familyID)
}

func (s *Service) GetItemsByFamilyID(familyID int) ([]*entities.Item, error) {
    return s.repo.GetByFamilyID(familyID)
}

func (s *Service) UpdateItem(id int, familyID int, req *UpdateItemRequest) (*entities.Item, error) {
    item := &entities.Item{
        ID:          id,
        Name:        req.Name,
        Description: req.Description,
        Quantity:    req.Quantity,
        ContainerID: req.ContainerID,
        FamilyID:    familyID,
        UpdatedAt:   time.Now().UTC(),
    }

    if req.Tags != nil {
        item.Tags = make([]entities.Tag, len(req.Tags))
        for i, tagID := range req.Tags {
            item.Tags[i] = entities.Tag{
                ID:       tagID,
                FamilyID: familyID,
            }
        }
    }

    if err := s.repo.Update(item); err != nil {
        return nil, fmt.Errorf("failed to update item: %v", err)
    }

    return s.repo.GetByID(id, familyID)
}

func (s *Service) AddItemImage(itemID int, familyID int, url string) error {
    displayOrder := 0
    item, err := s.repo.GetByID(itemID, familyID)
    if err == nil {
        displayOrder = len(item.Images)
    }
    
    return s.repo.AddItemImage(itemID, familyID, url, displayOrder)
}

func (s *Service) DeleteItemImage(itemID int, familyID int, url string) error {
    return s.repo.DeleteItemImage(itemID, familyID, url)
}

func (s *Service) DeleteItem(id int, familyID int, deletedBy int) error {
    return s.repo.Delete(id, familyID, deletedBy)
}

func (s *Service) RestoreItem(id int, familyID int) error {
    if err := s.repo.RestoreDeleted(id, familyID); err != nil {
        return fmt.Errorf("failed to restore item: %v", err)
    }
    return nil
}