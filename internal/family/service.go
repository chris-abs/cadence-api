package family

import (
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/email"
	"github.com/chrisabs/cadence/internal/models"
)

type Service struct {
	repo *Repository
	profileService interface {
		GetUserByID(id int) (*models.Profile, error)
	}
	emailService *email.Service
}

func NewService(
	repo *Repository,
	profileService interface {
		GetUserByID(id int) (*models.Profile, error)
	},
) *Service {
	emailService, err := email.NewService()
	if err != nil {
		fmt.Printf("Failed to initialize email service: %v\n", err)
	}

	return &Service{
		repo: repo,
		profileService: profileService,
		emailService: emailService,
	}
}

func (s *Service) CreateFamily(req *CreateFamilyRequest, ownerID int) (*models.Family, error) {
    family := &models.Family{
        Name:    req.Name,
        Status:  models.FamilyStatusActive,
    }

    if err := s.repo.Create(family); err != nil {
        return nil, fmt.Errorf("failed to create family: %v", err)
    }

    return family, nil
}

func (s *Service) GetFamily(id int) (*models.Family, error) {
	return s.repo.GetByID(id)
}

func (s *Service) UpdateFamily(familyID int, req *UpdateFamilyRequest) (*models.Family, error) {
    family, err := s.repo.GetByID(familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get family: %v", err)
    }
    
    family.Name = req.Name
    family.Status = req.Status
    family.UpdatedAt = time.Now().UTC()
    
    if err := s.repo.Update(family); err != nil {
        return nil, fmt.Errorf("failed to update family: %v", err)
    }
    
    return family, nil
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
			moduleFound = true
			break
		}
	}

	if !moduleFound {
		family.Modules = append(family.Modules, models.Module{
			ID:        req.ModuleID,
			IsEnabled: req.IsEnabled,
		})
	}

	if err := s.repo.Update(family); err != nil {
		return fmt.Errorf("failed to update family modules: %v", err)
	}

	return nil
}

func (s *Service) HasModulePermission(familyID int, profileRole models.ProfileRole, moduleID models.ModuleID, permission models.Permission) (bool, error) {
	family, err := s.repo.GetByID(familyID)
	if err != nil {
		return false, fmt.Errorf("failed to get family: %v", err)
	}

	if family.Status != models.FamilyStatusActive {
		return false, nil
	}

	permissions := map[models.ModuleID]map[models.ProfileRole][]models.Permission{
		"storage": {
			"PARENT": {models.PermissionRead, models.PermissionWrite, models.PermissionManage},
			"CHILD":  {models.PermissionRead},
		},
		"chores": {
			"PARENT": {models.PermissionRead, models.PermissionWrite, models.PermissionManage},
			"CHILD":  {models.PermissionRead, models.PermissionWrite},
		},
		"meals": {
			"PARENT": {models.PermissionRead, models.PermissionWrite, models.PermissionManage},
			"CHILD":  {models.PermissionRead},
		},
		"services": {
			"PARENT": {models.PermissionRead, models.PermissionWrite, models.PermissionManage},
			"CHILD":  {models.PermissionRead},
		},
	}

	modulePermissions, ok := permissions[moduleID]
	if !ok {
		return false, nil
	}

	rolePermissions, ok := modulePermissions[profileRole]
	if !ok {
		return false, nil
	}

	isEnabled, err := s.IsModuleEnabled(familyID, moduleID)
	if err != nil {
		return false, err
	}

	if !isEnabled {
		return false, nil
	}

	for _, p := range rolePermissions {
		if p == permission {
			return true, nil
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

func (s *Service) DeleteFamily(id int, deletedBy int) error {
    if err := s.repo.Delete(id, deletedBy); err != nil {
        return fmt.Errorf("failed to delete family: %v", err)
    }
    return nil
}

func (s *Service) RestoreFamily(id int) error {
    if err := s.repo.RestoreFamily(id); err != nil {
        return fmt.Errorf("failed to restore family: %v", err)
    }
    return nil
}