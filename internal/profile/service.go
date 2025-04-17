package profile

import (
	"fmt"
	"mime/multipart"
	"regexp"
	"time"

	"github.com/chrisabs/cadence/internal/cloud"
	"github.com/chrisabs/cadence/internal/models"
	"github.com/golang-jwt/jwt"
)

type Service struct {
	repo      *Repository
	jwtSecret string
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) GenerateProfileJWT(familyID, profileID int, role models.ProfileRole, isOwner bool) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["familyId"] = familyID
	claims["profileId"] = profileID
	claims["role"] = string(role)
	claims["isOwner"] = isOwner
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	return token.SignedString([]byte(s.jwtSecret))
}


func (s *Service) CreateProfile(familyID int, req *CreateProfileRequest) (*models.Profile, error) {
    existingProfiles, err := s.repo.GetByFamilyID(familyID)
    if err != nil {
        return nil, fmt.Errorf("error checking existing profiles: %v", err)
    }

    isOwner := len(existingProfiles) == 0

    if isOwner && req.Role != models.RoleParent {
        return nil, fmt.Errorf("owner profile must be a parent")
    }

    profile := &models.Profile{
        FamilyID:  familyID,
        Name:      req.Name,
        Role:      req.Role,
        Pin:       req.Pin,
        ImageURL:  req.ImageURL,
        IsOwner:   isOwner, 
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
    }

    if err := s.repo.Create(profile); err != nil {
        return nil, fmt.Errorf("failed to create profile: %v", err)
    }

    return s.repo.GetByID(profile.ID)
}

func (s *Service) GetProfileByID(id int) (*models.Profile, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetProfilesByFamilyID(familyID int) ([]*models.Profile, error) {
	return s.repo.GetByFamilyID(familyID)
}

func (s *Service) UpdateProfile(id int, familyID int, req *UpdateProfileRequest, imageFile *multipart.FileHeader) (*models.Profile, error) {
	profile, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %v", err)
	}

	if profile.FamilyID != familyID {
		return nil, fmt.Errorf("profile does not belong to this family")
	}

	if req.Name != "" {
		profile.Name = req.Name
	}

	if req.Role != "" {
		if profile.IsOwner && req.Role != models.RoleParent {
			return nil, fmt.Errorf("cannot change role of owner profile")
		}
		profile.Role = req.Role
	}

	if req.Pin != nil {
		if profile.HasPin && profile.Pin != "" {
			if req.CurrentPin == "" {
				return nil, fmt.Errorf("current PIN required to change PIN")
			}
			if req.CurrentPin != profile.Pin {
				return nil, fmt.Errorf("invalid current PIN")
			}
		}
	
		if *req.Pin == "" {
			profile.Pin = ""
			profile.HasPin = false
		} else {
			if len(*req.Pin) != 6 || !regexp.MustCompile(`^\d{6}$`).MatchString(*req.Pin) {
				return nil, fmt.Errorf("PIN must be exactly 6 digits")
			}
			profile.Pin = *req.Pin
			profile.HasPin = true
		}
	}

	if imageFile != nil {
		s3Handler, err := cloud.NewS3Handler()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage: %v", err)
		}

		imageURL, err := s3Handler.UploadFile(imageFile, fmt.Sprintf("profiles/%d", id))
		if err != nil {
			return nil, fmt.Errorf("failed to upload image: %v", err)
		}
		profile.ImageURL = imageURL
	} else if req.ImageURL != "" {
		profile.ImageURL = req.ImageURL
	}

	profile.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(profile); err != nil {
		return nil, fmt.Errorf("failed to update profile: %v", err)
	}

	return s.repo.GetByID(profile.ID)
}

func (s *Service) DeleteProfile(id int, familyID int, deletedBy int) error {
	profile, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("profile not found: %v", err)
	}

	if profile.FamilyID != familyID {
		return fmt.Errorf("profile does not belong to this family")
	}

	if profile.IsOwner {
		return fmt.Errorf("cannot delete the owner profile")
	}

	return s.repo.Delete(id, familyID, deletedBy)
}

func (s *Service) RestoreProfile(id int, familyID int) error {
	return s.repo.Restore(id, familyID)
}

func (s *Service) VerifyPin(familyID int, profileID int, pin string) (*ProfileResponse, error) {
	profile, err := s.repo.GetByID(profileID)
	if err != nil {
		return nil, fmt.Errorf("profile not found")
	}

	if profile.FamilyID != familyID {
		return nil, fmt.Errorf("profile does not belong to this family")
	}

	if profile.HasPin && (pin == "" || profile.Pin != pin) {
		return nil, fmt.Errorf("invalid PIN")
	}

	token, err := s.GenerateProfileJWT(familyID, profileID, profile.Role, profile.IsOwner)
	if err != nil {
		return nil, fmt.Errorf("error generating token")
	}

	return &ProfileResponse{
		Token:   token,
		Profile: *profile,
	}, nil
}

func (s *Service) SelectProfile(familyID int, req *SelectProfileRequest) (*ProfileResponse, error) {
	return s.VerifyPin(familyID, req.ProfileID, req.Pin)
}

func (s *Service) GetOwnerProfile(familyID int) (*models.Profile, error) {
	return s.repo.GetOwnerProfile(familyID)
}