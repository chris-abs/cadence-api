package container

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chrisabs/storage/internal/item"
	"github.com/chrisabs/storage/pkg/utils"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateContainer(req *CreateContainerRequest) (*Container, error) {
	containerID := rand.Intn(10000)
	qrString, qrImage, err := utils.GenerateQRCode(containerID)
	if err != nil {
		qrString = fmt.Sprintf("STQRAGE-CONTAINER-%d", containerID)
		qrImage = ""
	}

	container := &Container{
		ID:          containerID,
		Name:        req.Name,
		QRCode:      qrString,
		QRCodeImage: qrImage,
		Number:      rand.Intn(1000),
		Location:    req.Location,
		Items:       []item.Item{},
		UserID:      rand.Intn(10000),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.repo.Create(container, req.ItemIDs); err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	return s.repo.GetByID(container.ID)
}

func (s *Service) GetContainerByID(id int) (*Container, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetContainerByQR(qrCode string) (*Container, error) {
	return s.repo.GetByQR(qrCode)
}

func (s *Service) GetAllContainers() ([]*Container, error) {
	return s.repo.GetAll()
}

func (s *Service) UpdateContainer(id int, req *UpdateContainerRequest) (*Container, error) {
	container, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("container not found: %v", err)
	}

	container.Name = req.Name
	container.Location = req.Location

	if err := s.repo.Update(container, req.ItemIDs); err != nil {
		return nil, fmt.Errorf("failed to update container: %v", err)
	}

	return container, nil
}

func (s *Service) UpdateContainerItems(containerID int, itemIDs []int) error {
	return s.repo.UpdateContainerItems(containerID, itemIDs)
}

func (s *Service) DeleteContainer(id int) error {
	return s.repo.Delete(id)
}
