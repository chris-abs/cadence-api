package user

import (
	"fmt"
	"mime/multipart"
	"time"

	"github.com/chrisabs/cadence/internal/cloud"
	"github.com/chrisabs/cadence/internal/models"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo          *Repository
	jwtSecret     string
	familyService interface {
		ValidateInvite(token string) (*models.FamilyInvite, error)
		DeleteInvite(id int) error
	}
	membershipService interface {
		CreateMembership(userID, familyID int, role models.UserRole, isOwner bool) (*models.FamilyMembership, error)
		GetActiveMembershipForUser(userID int) (*models.FamilyMembership, error)
	}
}

func NewService(
	repo *Repository,
	familyService interface {
		ValidateInvite(token string) (*models.FamilyInvite, error)
		DeleteInvite(id int) error
	},
	jwtSecret string,
) *Service {
	return &Service{
		repo:          repo,
		familyService: familyService,
		jwtSecret:     jwtSecret,
	}
}

func (s *Service) SetMembershipService(membershipService interface {
	CreateMembership(userID, familyID int, role models.UserRole, isOwner bool) (*models.FamilyMembership, error)
	GetActiveMembershipForUser(userID int) (*models.FamilyMembership, error)
}) {
	s.membershipService = membershipService
}

func (s *Service) generateJWT(userID int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["userId"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) CreateUser(req *CreateUserRequest) (*models.User, error) {
	existingUser, err := s.repo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	tx, err := s.repo.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	user := &models.User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		ImageURL:  req.ImageURL,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateTx(tx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return s.repo.GetByID(user.ID)
}

func (s *Service) Login(req *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	token, err := s.generateJWT(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	response := &AuthResponse{
		Token: token,
		User:  *user,
	}

	if s.membershipService != nil {
		membership, err := s.membershipService.GetActiveMembershipForUser(user.ID)
		if err == nil && membership != nil {
			response.FamilyID = &membership.FamilyID
			response.Role = &membership.Role
		} else {
			response.FamilyID = nil
			response.Role = nil
		}
	}

	return response, nil
}

func (s *Service) AcceptInvite(req *AcceptInviteRequest) (*models.User, error) {
	invite, err := s.familyService.ValidateInvite(req.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid invite: %v", err)
	}

	existingUser, err := s.repo.GetByEmail(invite.Email)
	if err != nil && err.Error() != "user not found" {
		return nil, fmt.Errorf("error checking existing user: %v", err)
	}

	var user *models.User

	if existingUser == nil {
		if req.Password == "" {
			return nil, fmt.Errorf("password required for new user")
		}

		user = &models.User{
			Email:    invite.Email,
			Password: req.Password,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		tx, err := s.repo.BeginTx()
		if err != nil {
			return nil, fmt.Errorf("failed to start transaction: %v", err)
		}
		defer tx.Rollback()

		if err := s.repo.CreateTx(tx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %v", err)
		}

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %v", err)
		}

		user, err = s.repo.GetByID(user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get created user: %v", err)
		}
	} else {
		user = existingUser
	}

	if s.membershipService != nil {
		_, err = s.membershipService.CreateMembership(user.ID, invite.FamilyID, invite.Role, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create membership: %v", err)
		}
	}

	if err := s.familyService.DeleteInvite(invite.ID); err != nil {
		fmt.Printf("failed to delete used invite: %v\n", err)
	}

	return user, nil
}

func (s *Service) GetUserByID(id int) (*models.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) GetAllUsers() ([]*models.User, error) {
	return s.repo.GetAll()
}

func (s *Service) UpdateUser(id int, firstName, lastName string, imageFile *multipart.FileHeader) (*models.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedAt = time.Now().UTC()

	if imageFile != nil {
		s3Handler, err := cloud.NewS3Handler()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage: %v", err)
		}

		imageURL, err := s3Handler.UploadFile(imageFile, fmt.Sprintf("users/%d", id))
		if err != nil {
			return nil, fmt.Errorf("failed to upload image: %v", err)
		}
		user.ImageURL = imageURL
	}

	if err := s.repo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	return s.GetUserByID(id)
}

func (s *Service) DeleteUser(id int) error {
	return s.repo.Delete(id)
}