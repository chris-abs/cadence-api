package container

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chrisabs/storage/internal/models"
	"github.com/chrisabs/storage/pkg/utils"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateContainer(userID int, req *CreateContainerRequest) (*models.Container, error) {
    containerID := rand.Intn(10000)
    qrString, qrImage, err := utils.GenerateQRCode(containerID)
    if err != nil {
        qrString = fmt.Sprintf("STQRAGE-CONTAINER-%d", containerID)
        qrImage = ""
    }

    container := &models.Container{
        ID:          containerID,
        Name:        req.Name,
        QRCode:      qrString,
        QRCodeImage: qrImage,
        Number:      rand.Intn(1000),
        Location:    req.Location,
        UserID:      userID,
        WorkspaceID: nil, 
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
    }

	if req.WorkspaceID != nil {
        container.WorkspaceID = req.WorkspaceID
    }

	if err := s.repo.Create(container, req.Items); err != nil {
		return nil, fmt.Errorf("failed to create container with items: %v", err)
	}

	return s.repo.GetByID(container.ID)
}
func (s *Service) GetContainerByID(id int) (*models.Container, error) {
	container, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("error getting container: %v", err)
	}
	return container, nil
}

func (s *Service) GetContainersByUserID(userID int) ([]*models.Container, error) {
	return s.repo.GetByUserID(userID)
}

func (s *Service) UpdateContainer(id int, req *UpdateContainerRequest) (*models.Container, error) {
	container, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("container not found: %v", err)
	}

	container.Name = req.Name
	container.Location = req.Location
	container.WorkspaceID = &req.WorkspaceID
	container.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(container); err != nil {
		return nil, fmt.Errorf("failed to update container: %v", err)
	}

	return container, nil
}

func (s *Service) DeleteContainer(id int) error {
	return s.repo.Delete(id)
}

func (s *Service) GetContainerByQR(qrCode string) (*models.Container, error) {
	return s.repo.GetByQR(qrCode)
}
