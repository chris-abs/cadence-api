package container

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

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
		return nil, fmt.Errorf("failed to generate QR code: %v", err)
	}

	container := &Container{
		ID:          containerID,
		Name:        req.Name,
		QRCode:      qrString,
		QRCodeImage: qrImage,
		Number:      rand.Intn(1000),
		Location:    req.Location,
		UserID:      rand.Intn(10000), // TODO This should be replaced with actual user ID
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.repo.Create(container); err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	return container, nil
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

func (s *Service) UpdateContainer(id int, updates map[string]interface{}) (*Container, error) {
	container, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("container not found: %v", err)
	}

	// Use reflection to update fields
	containerValue := reflect.ValueOf(container).Elem()
	for key, value := range updates {
		field := containerValue.FieldByNameFunc(func(fieldName string) bool {
			return fieldName == key
		})
		if field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(value))
		}
	}

	container.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(container); err != nil {
		return nil, fmt.Errorf("failed to update container: %v", err)
	}

	return container, nil
}

func (s *Service) DeleteContainer(id int) error {
	return s.repo.Delete(id)
}
