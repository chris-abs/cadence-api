package family

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/chrisabs/storage/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateFamily(req *CreateFamilyRequest, ownerID int) (*models.Family, error) {
	family := &models.Family{
		Name:    req.Name,
		OwnerID: ownerID,
		Status:  models.FamilyStatusActive, 
	}

	if err := s.repo.Create(family); err != nil {
		return nil, fmt.Errorf("failed to create family: %v", err)
	}

	return family, nil
}

func (s *Service) CreateInvite(req *CreateInviteRequest) (*models.FamilyInvite, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	invite := &models.FamilyInvite{
		FamilyID:  req.FamilyID,
		Email:     req.Email,
		Role:      req.Role,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
	}

	if err := s.repo.CreateInvite(invite); err != nil {
		return nil, fmt.Errorf("failed to create invite: %v", err)
	}

	return invite, nil
}

func (s *Service) ValidateInvite(token string) (*models.FamilyInvite, error) {
	invite, err := s.repo.GetInviteByToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired invite: %v", err)
	}

	return invite, nil
}

func (s *Service) GetFamily(id int) (*models.Family, error) {
	return s.repo.GetByID(id)
}

func (s *Service) UpdateModuleSettings(familyID int, req *UpdateModuleRequest) error {
	family, err := s.repo.GetByID(familyID)
	if err != nil {
		return fmt.Errorf("failed to get family: %v", err)
	}

	if family.Status != models.FamilyStatusActive {
		return fmt.Errorf("family is not active") 
	}

	moduleFound := false
	for i, module := range family.Modules {
		if module.ID == req.ModuleID {
			family.Modules[i].IsEnabled = req.IsEnabled
			family.Modules[i].Settings.Permissions = req.Permissions
			moduleFound = true
			break
		}
	}

	if !moduleFound {
		family.Modules = append(family.Modules, models.Module{
			ID:        req.ModuleID,
			IsEnabled: req.IsEnabled,
			Settings: models.ModuleSettings{
				Permissions: req.Permissions,
			},
		})
	}

	if err := s.repo.Update(family); err != nil {
		return fmt.Errorf("failed to update family modules: %v", err)
	}

	return nil
}

func (s *Service) HasModulePermission(familyID int, userRole models.UserRole, moduleID models.ModuleID, permission models.Permission) (bool, error) {
	family, err := s.repo.GetByID(familyID)
	if err != nil {
		return false, fmt.Errorf("failed to get family: %v", err)
	}

	if family.Status != models.FamilyStatusActive {
		return false, nil 
	}

	for _, module := range family.Modules {
		if module.ID == moduleID && module.IsEnabled {
			permissions, exists := module.Settings.Permissions[userRole]
			if !exists {
				return false, nil
			}

			for _, p := range permissions {
				if p == permission {
					return true, nil
				}
			}
			return false, nil
		}
	}

	return false, nil
}

func (s *Service) IsModuleEnabled(familyID int, moduleID models.ModuleID) (bool, error) {
	family, err := s.repo.GetByID(familyID)
	if err != nil {
		return false, fmt.Errorf("failed to get family: %v", err)
	}

	if family.Status != models.FamilyStatusActive {
		return false, nil 
	}

	for _, module := range family.Modules {
		if module.ID == moduleID {
			return module.IsEnabled, nil
		}
	}

	return false, nil
}

func (s *Service) GetFamilyModules(familyID int) ([]models.Module, error) {
	family, err := s.repo.GetByID(familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family: %v", err)
	}

	return family.Modules, nil
}

func (s *Service) DeleteInvite(id int) error {
	return s.repo.DeleteInvite(id)
}

func (s *Service) DeactivateFamily(familyID int) error {
	family, err := s.repo.GetByID(familyID)
	if err != nil {
		return fmt.Errorf("failed to get family: %v", err)
	}

	family.Status = models.FamilyStatusInactive
	if err := s.repo.Update(family); err != nil {
		return fmt.Errorf("failed to update family status: %v", err)
	}

	return nil
}
