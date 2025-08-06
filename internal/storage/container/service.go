package container

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chrisabs/cadence/internal/models"
	"github.com/chrisabs/cadence/internal/storage/entities"
	"github.com/chrisabs/cadence/pkg/utils"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateContainer(profileId models.ProfileID, familyID models.FamilyID, req *CreateContainerRequest) (*entities.Container, error) {
    
    containerID := models.NewContainerID()
    
    qrString, qrImage, err := utils.GenerateQRCode(containerID.String())
    if err != nil {
        qrString = containerID.String()
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
        ProfileID:   profileId,
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

func (s *Service) GetContainerByID(id models.ContainerID, familyID models.FamilyID) (*entities.Container, error) {
    container, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, fmt.Errorf("error getting container: %v", err)
    }
    return container, nil
}

func (s *Service) GetContainersByFamilyID(familyID models.FamilyID) ([]*entities.Container, error) {
    return s.repo.GetByFamilyID(familyID)
}

func (s *Service) UpdateContainer(id models.ContainerID, familyID models.FamilyID, req *UpdateContainerRequest) (*entities.Container, error) {
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

func (s *Service) GetContainerByQR(qrCode string, familyID models.FamilyID) (*entities.Container, error) {
    return s.repo.GetByQR(qrCode, familyID)
}

func (s *Service) DeleteContainer(id models.ContainerID, familyID models.FamilyID, deletedBy models.ProfileID) error {
    if err := s.repo.Delete(id, familyID, deletedBy); err != nil {
        return fmt.Errorf("failed to delete container: %v", err)
    }
    return nil
}

func (s *Service) RestoreContainer(id models.ContainerID, familyID models.FamilyID) error {
    if err := s.repo.RestoreDeleted(id, familyID); err != nil {
        return fmt.Errorf("failed to restore container: %v", err)
    }
    return nil
}