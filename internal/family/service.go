package family

import (
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/models"
	"github.com/chrisabs/cadence/internal/profile"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo         *Repository
	jwtSecret    string
	profileService interface {
		CreateProfile(familyID int, req *profile.CreateProfileRequest) (*profile.Profile, error)
		GetProfilesByFamilyID(familyID int) ([]*profile.Profile, error)
	}
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) SetProfileService(profileService interface {
	CreateProfile(familyID int, req *profile.CreateProfileRequest) (*profile.Profile, error)
	GetProfilesByFamilyID(familyID int) ([]*profile.Profile, error)
}) {
	s.profileService = profileService
}

func (s *Service) GenerateFamilyJWT(familyID int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["familyId"] = familyID
	claims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix() // 7 days

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) Register(req *RegisterRequest) (*FamilyAuthResponse, error) {
	existingFamily, err := s.repo.GetByEmail(req.Email)
	if err == nil && existingFamily != nil {
		return nil, fmt.Errorf("email already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}

	family := &FamilyAccount{
		Email:      req.Email,
		Password:   string(hashedPassword),
		FamilyName: req.FamilyName,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := s.repo.Create(family); err != nil {
		return nil, fmt.Errorf("failed to create family account: %v", err)
	}

	settings := &FamilySettings{
		FamilyID: family.ID,
		Modules: []models.Module{
			{ID: models.ModuleStorage, IsEnabled: true},
			{ID: models.ModuleChores, IsEnabled: false},
			{ID: models.ModuleMeals, IsEnabled: false},
			{ID: models.ModuleServices, IsEnabled: false},
		},
		Status:    models.FamilyStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateSettings(settings); err != nil {
		return nil, fmt.Errorf("failed to create family settings: %v", err)
	}

	var profiles []profile.Profile
	if s.profileService != nil {
		ownerProfile, err := s.profileService.CreateProfile(family.ID, &profile.CreateProfileRequest{
			Name:  req.OwnerName,
			Role:  models.RoleParent,
			Pin:   "",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create owner profile: %v", err)
		}
		
		profiles = append(profiles, *ownerProfile)
	}

	token, err := s.GenerateFamilyJWT(family.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return &FamilyAuthResponse{
		Token:    token,
		Family:   *family,
		Profiles: profiles,
	}, nil
}

func (s *Service) Login(req *LoginRequest) (*FamilyAuthResponse, error) {
	family, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(family.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	var profiles []profile.Profile
	if s.profileService != nil {
		profilesPtr, err := s.profileService.GetProfilesByFamilyID(family.ID)
		if err == nil {
			for _, p := range profilesPtr {
				profiles = append(profiles, *p)
			}
		}
	}

	token, err := s.GenerateFamilyJWT(family.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return &FamilyAuthResponse{
		Token:    token,
		Family:   *family,
		Profiles: profiles,
	}, nil
}

func (s *Service) GetFamilyByID(id int) (*FamilyAccount, error) {
	return s.repo.GetByID(id)
}

func (s *Service) UpdateFamily(id int, req *UpdateFamilyRequest) (*FamilyAccount, error) {
	family, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("family not found: %v", err)
	}

	family.FamilyName = req.FamilyName
	family.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(family); err != nil {
		return nil, fmt.Errorf("failed to update family: %v", err)
	}

	return family, nil
}

func (s *Service) GetFamilySettings(familyID int) (*FamilySettings, error) {
	return s.repo.GetSettings(familyID)
}

func (s *Service) UpdateModule(familyID int, req *UpdateModuleRequest) error {
	return s.repo.UpdateModule(familyID, req.ModuleID, req.IsEnabled)
}

func (s *Service) DeleteFamily(id int, deletedBy int) error {
	return s.repo.Delete(id, deletedBy)
}

func (s *Service) RestoreFamily(id int) error {
	return s.repo.Restore(id)
}

func (s *Service) IsModuleEnabled(familyID int, moduleID models.ModuleID) (bool, error) {
	return s.repo.IsModuleEnabled(familyID, moduleID)
}

func (s *Service) HasModulePermission(familyID int, role models.ProfileRole, moduleID models.ModuleID, permission models.Permission) (bool, error) {
	isEnabled, err := s.IsModuleEnabled(familyID, moduleID)
	if err != nil {
		return false, err
	}

	if !isEnabled {
		return false, nil
	}

	permissions := map[models.ModuleID]map[models.ProfileRole][]models.Permission{
		models.ModuleStorage: {
			models.RoleParent: {models.PermissionRead, models.PermissionWrite, models.PermissionManage},
			models.RoleChild:  {models.PermissionRead},
		},
		models.ModuleChores: {
			models.RoleParent: {models.PermissionRead, models.PermissionWrite, models.PermissionManage},
			models.RoleChild:  {models.PermissionRead, models.PermissionWrite},
		},
		models.ModuleMeals: {
			models.RoleParent: {models.PermissionRead, models.PermissionWrite, models.PermissionManage},
			models.RoleChild:  {models.PermissionRead},
		},
		models.ModuleServices: {
			models.RoleParent: {models.PermissionRead, models.PermissionWrite, models.PermissionManage},
			models.RoleChild:  {models.PermissionRead},
		},
	}

	modulePermissions, ok := permissions[moduleID]
	if !ok {
		return false, nil
	}

	rolePermissions, ok := modulePermissions[role]
	if !ok {
		return false, nil
	}

	for _, p := range rolePermissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}