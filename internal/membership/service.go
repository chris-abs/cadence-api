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

func (s *Service) CreateMembership(profileId, familyID int, role models.UserRole, isOwner bool) (*models.FamilyMembership, error) {
	membership := &models.FamilyMembership{
		profileId:   profileId,
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

func (s *Service) GetMembershipsByprofileId(profileId int) ([]*models.FamilyMembership, error) {
	return s.repo.GetByprofileId(profileId)
}

func (s *Service) GetActiveMembershipForUser(profileId int) (*models.FamilyMembership, error) {
	return s.repo.GetActiveMembershipForUser(profileId)
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

func (s *Service) DeleteMembership(id int, deletedBy int) error {
    return s.repo.Delete(id, deletedBy)
}

func (s *Service) RestoreMembership(id int) error {
    if err := s.repo.RestoreMembership(id); err != nil {
        return fmt.Errorf("failed to restore membership: %v", err)
    }
    return nil
}

func (s *Service) HasUserRole(profileId int, familyID int, role models.UserRole) (bool, error) {
	memberships, err := s.repo.GetByprofileId(profileId)
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

func (s *Service) IsUserFamilyOwner(profileId int, familyID int) (bool, error) {
	memberships, err := s.repo.GetByprofileId(profileId)
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

func (s *Service) IsUserInFamily(profileId, familyID int) (bool, error) {
	memberships, err := s.repo.GetByprofileId(profileId)
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

func (s *Service) GetUserFamilyAndRole(profileId int) (*int, *models.UserRole, error) {
	membership, err := s.GetActiveMembershipForUser(profileId)
	if err != nil {
		return nil, nil, err
	}
	
	familyID := membership.FamilyID
	role := membership.Role
	
	return &familyID, &role, nil
}