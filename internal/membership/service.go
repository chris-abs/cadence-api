package membership

import (
	"fmt"

	"github.com/chrisabs/cadence/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateMembership(userID, familyID int, role models.UserRole, isOwner bool) (*models.FamilyMembership, error) {
	membership := &models.FamilyMembership{
		UserID:   userID,
		FamilyID: familyID,
		Role:     role,
		IsOwner:  isOwner,
	}

	if err := s.repo.Create(membership); err != nil {
		return nil, fmt.Errorf("failed to create membership: %v", err)
	}

	return membership, nil
}

func (s *Service) GetMembershipByID(id int) (*models.FamilyMembership, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetMembershipsByUserID(userID int) ([]*models.FamilyMembership, error) {
	return s.repo.GetByUserID(userID)
}

func (s *Service) GetActiveMembershipForUser(userID int) (*models.FamilyMembership, error) {
	return s.repo.GetActiveMembershipForUser(userID)
}

func (s *Service) GetMembershipsByFamilyID(familyID int) ([]*models.FamilyMembership, error) {
	return s.repo.GetByFamilyID(familyID)
}

func (s *Service) GetFamilyOwner(familyID int) (*models.FamilyMembership, error) {
	return s.repo.GetFamilyOwner(familyID)
}

func (s *Service) UpdateMembership(id int, role models.UserRole, isOwner bool) (*models.FamilyMembership, error) {
	membership, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("membership not found: %v", err)
	}

	membership.Role = role
	membership.IsOwner = isOwner

	if err := s.repo.Update(membership); err != nil {
		return nil, fmt.Errorf("failed to update membership: %v", err)
	}

	return membership, nil
}

func (s *Service) DeleteMembership(id int) error {
	return s.repo.Delete(id)
}

func (s *Service) HasUserRole(userID int, familyID int, role models.UserRole) (bool, error) {
	memberships, err := s.repo.GetByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("error getting memberships: %v", err)
	}

	for _, m := range memberships {
		if m.FamilyID == familyID && m.Role == role {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) IsUserFamilyOwner(userID int, familyID int) (bool, error) {
	memberships, err := s.repo.GetByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("error getting memberships: %v", err)
	}

	for _, m := range memberships {
		if m.FamilyID == familyID && m.IsOwner {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) IsUserInFamily(userID, familyID int) (bool, error) {
	memberships, err := s.repo.GetByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("error getting memberships: %v", err)
	}

	for _, m := range memberships {
		if m.FamilyID == familyID {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) GetUserFamilyAndRole(userID int) (*int, *models.UserRole, error) {
	membership, err := s.GetActiveMembershipForUser(userID)
	if err != nil {
		return nil, nil, err
	}
	
	familyID := membership.FamilyID
	role := membership.Role
	
	return &familyID, &role, nil
}