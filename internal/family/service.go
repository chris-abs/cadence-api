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
	userService interface {
		GetUserByID(id int) (*models.User, error)
		UpdateFamily(user *models.User) error
	}
}

func NewService(repo *Repository, userService interface {
	GetUserByID(id int) (*models.User, error)
	UpdateFamily(user *models.User) error
},
) *Service {
	return &Service{
		repo: repo,
		userService: userService,
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

    user, err := s.userService.GetUserByID(ownerID)
    if err != nil {
        return family, fmt.Errorf("family created but failed to get user: %v", err)
    }

    parentRole := models.RoleParent
    user.FamilyID = &family.ID
    user.Role = &parentRole

    if err := s.userService.UpdateFamily(user); err != nil {
        return family, fmt.Errorf("family created but failed to update user family: %v", err)
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

func (s *Service) HasModulePermission(familyID int, userRole models.UserRole, moduleID models.ModuleID, permission models.Permission) (bool, error) {
	family, err := s.repo.GetByID(familyID)
	if err != nil {
		return false, fmt.Errorf("failed to get family: %v", err)
	}

	if family.Status != models.FamilyStatusActive {
		return false, nil
	}

	permissions := map[models.ModuleID]map[models.UserRole][]models.Permission{
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

	rolePermissions, ok := modulePermissions[userRole]
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

func (s *Service) JoinFamily(userID int, req *JoinFamilyRequest) (*models.User, error) {
	invite, err := s.ValidateInvite(req.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid invite: %v", err)
	}

	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	familyID := invite.FamilyID
	role := invite.Role
	user.FamilyID = &familyID
	user.Role = &role

	if err := s.userService.UpdateFamily(user); err != nil {
		return nil, fmt.Errorf("failed to update user family: %v", err)
	}

	if err := s.DeleteInvite(invite.ID); err != nil {
		fmt.Printf("failed to delete used invite: %v\n", err) 
	}

	return user, nil
}
