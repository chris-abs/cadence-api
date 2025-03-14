package container

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chrisabs/cadence/internal/storage/entities"
	"github.com/chrisabs/cadence/pkg/utils"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateContainer(profileId int, familyID int, req *CreateContainerRequest) (*entities.Container, error) {
    containerID := rand.Intn(10000)
    qrString, qrImage, err := utils.GenerateQRCode(containerID)
    if err != nil {
        qrString = fmt.Sprintf("STORAGE-CONTAINER-%d", containerID)
        qrImage = ""
    }

    container := &entities.Container{
        ID:          containerID,
        Name:        req.Name,
        Description: req.Description,
        QRCode:      qrString,
        QRCodeImage: qrImage,
        Number:      rand.Intn(1000),
        Location:    req.Location,
        profileId:      profileId,
        FamilyID:    familyID,        
        WorkspaceID: req.WorkspaceID,
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
    }

    if err := s.repo.Create(container, req.Items); err != nil {
        return nil, fmt.Errorf("failed to create container with items: %v", err)
    }

    return s.repo.GetByID(container.ID, familyID)
}

func (s *Service) GetContainerByID(id int, familyID int) (*entities.Container, error) {
    container, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, fmt.Errorf("error getting container: %v", err)
    }
    return container, nil
}

func (s *Service) GetContainersByFamilyID(familyID int) ([]*entities.Container, error) {
    return s.repo.GetByFamilyID(familyID)
}

func (s *Service) UpdateContainer(id int, familyID int, req *UpdateContainerRequest) (*entities.Container, error) {
    container, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, fmt.Errorf("container not found: %v", err)
    }

    container.Name = req.Name
    container.Description = req.Description
    container.Location = req.Location
    container.WorkspaceID = req.WorkspaceID
    container.UpdatedAt = time.Now().UTC()

    if err := s.repo.Update(container); err != nil {
        return nil, fmt.Errorf("failed to update container: %v", err)
    }

    return container, nil
}

func (s *Service) GetContainerByQR(qrCode string, familyID int) (*entities.Container, error) {
    return s.repo.GetByQR(qrCode, familyID)
}

func (s *Service) DeleteContainer(id int, familyID int, deletedBy int) error {
    if err := s.repo.Delete(id, familyID, deletedBy); err != nil {
        return fmt.Errorf("failed to delete container: %v", err)
    }
    return nil
}

func (s *Service) RestoreContainer(id int, familyID int) error {
    if err := s.repo.RestoreDeleted(id, familyID); err != nil {
        return fmt.Errorf("failed to restore container: %v", err)
    }
    return nil
}