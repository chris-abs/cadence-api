package item

import (
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetItemByID(id int) (*Item, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetItemsByUserID(userID int) ([]*Item, error) {
	return s.repo.GetByUserID(userID)
}

func (s *Service) GetAllItems() ([]*Item, error) {
	return s.repo.GetAll()
}

func (s *Service) CreateItem(req *CreateItemRequest) (*Item, error) {
	item := &Item{
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Quantity:    req.Quantity,
		ContainerID: req.ContainerID,
	}

	itemID, err := s.repo.Create(item, req.TagIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %v", err)
	}

	return s.repo.GetByID(itemID)
}

func (s *Service) UpdateItem(id int, req *CreateItemRequest) (*Item, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("item not found: %v", err)
	}

	item.Name = req.Name
	item.Description = req.Description
	item.ImageURL = req.ImageURL
	item.Quantity = req.Quantity
	item.ContainerID = req.ContainerID

	if err := s.repo.Update(item, req.TagIDs); err != nil {
		return nil, fmt.Errorf("failed to update item: %v", err)
	}

	return s.repo.GetByID(id)
}

func (s *Service) DeleteItem(id int) error {
	return s.repo.Delete(id)
}
