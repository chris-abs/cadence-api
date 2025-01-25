package item

import (
	"fmt"
	"time"

	"github.com/chrisabs/storage/internal/models"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateItem(req *CreateItemRequest) (*models.Item, error) {
    if req.Name == "" {
        return nil, fmt.Errorf("item name is required")
    }

    item := &models.Item{
        Name:        req.Name,
        Description: req.Description,
        Quantity:    req.Quantity,
        ContainerID: req.ContainerID,
        Images:      []models.ItemImage{},
        Tags:        make([]models.Tag, 0),
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
    }

    createdItem, err := s.repo.Create(item, req.TagNames)
    if err != nil {
        return nil, fmt.Errorf("failed to create item: %v", err)
    }

    return createdItem, nil
}

func (s *Service) GetItemByID(id int) (*models.Item, error) {
    return s.repo.GetByID(id)
}

func (s *Service) GetItemsByUserID(userID int) ([]*models.Item, error) {
    return s.repo.GetByUserID(userID)
}

func (s *Service) UpdateItem(id int, req *UpdateItemRequest) (*models.Item, error) {
    item, err := s.repo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("item not found: %v", err)
    }

    item.Name = req.Name
    item.Description = req.Description
    item.Quantity = req.Quantity
    
    if req.ContainerID != nil {
        item.ContainerID = req.ContainerID
    } else {
        item.ContainerID = nil 
    }
    
    item.UpdatedAt = time.Now().UTC()

    if req.Tags != nil {
        item.Tags = make([]models.Tag, len(req.Tags))
        for i, tagID := range req.Tags {
            item.Tags[i] = models.Tag{ID: tagID}
        }
    } else {
        item.Tags = []models.Tag{} 
    }

    if err := s.repo.Update(item); err != nil {
        return nil, fmt.Errorf("failed to update item: %v", err)
    }

    return s.repo.GetByID(id)
}

func (s *Service) AddItemImage(itemID int, url string) error {
    item, err := s.repo.GetByID(itemID)
    if err != nil {
        return fmt.Errorf("item not found: %v", err)
    }

    displayOrder := len(item.Images)
    return s.repo.AddItemImage(itemID, url, displayOrder)
}

func (s *Service) DeleteItemImage(itemID int, url string) error {
    return s.repo.DeleteItemImage(itemID, url)
}

func (s *Service) DeleteItem(id int) error {
    return s.repo.Delete(id)
}