package material

import (
	"fmt"

	"github.com/chrisabs/cadence/internal/media/material/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateMaterial(profileID models.ProfileID, familyID models.FamilyID, req *CreateMaterialRequest) (*entities.Material, error) {
	if err := s.validateCreateMaterialRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	if req.Status == "" {
		req.Status = entities.StatusToWatch
	}

	if req.Priority == "" {
		req.Priority = entities.PriorityMedium
	}

	if req.Runtime == "" {
		req.Runtime = entities.RuntimeMedium
	}

	return s.repo.Create(profileID, familyID, req)
}

func (s *Service) GetMaterialByID(id models.MaterialID, profileID models.ProfileID) (*entities.Material, error) {
	if id == "" {
		return nil, fmt.Errorf("invalid material ID")
	}

	return s.repo.GetByID(id, profileID)
}

func (s *Service) SearchMaterial(familyID models.FamilyID, currentProfileID models.ProfileID, req *MaterialSearchRequest) (*MaterialSearchResponse, error) {
	if req.ProfileID == nil {
		req.ProfileID = &currentProfileID
	}

	if req.SortBy == "" {
		req.SortBy = "highest_rating"
	}

	if req.Status != "" {
		return s.repo.Search(familyID, req)
	} else {
		return s.repo.SearchAllColumns(familyID, req)
	}
}

func (s *Service) UpdateMaterial(id models.MaterialID, familyID models.FamilyID, profileID models.ProfileID, req *UpdateMaterialRequest) (*entities.Material, error) {
	if id == "" {
		return nil, fmt.Errorf("invalid material ID")
	}

	if err := s.validateUpdateMaterialRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	return s.repo.Update(id, familyID, profileID, req)
}

func (s *Service) UpdateMaterialStatus(id models.MaterialID, familyID models.FamilyID, profileID models.ProfileID, status entities.Status) error {
	if id == "" {
		return fmt.Errorf("invalid material ID")
	}

	if !s.isValidStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	return s.repo.UpdateStatus(id, familyID, profileID, status)
}

func (s *Service) UpdateMaterialReview(id models.MaterialID, familyID models.FamilyID, profileID models.ProfileID, reviewScore float64) error {
	if id == "" {
		return fmt.Errorf("invalid material ID")
	}

	if reviewScore < 0.0 || reviewScore > 10.0 {
		return fmt.Errorf("review score must be between 0.0 and 10.0")
	}

	return s.repo.UpdateReview(id, familyID, profileID, reviewScore)
}

func (s *Service) DeleteMaterial(id models.MaterialID, familyID models.FamilyID, profileID models.ProfileID, deletedBy models.ProfileID) error {
	if id == "" {
		return fmt.Errorf("invalid material ID")
	}

	return s.repo.Delete(id, familyID, profileID, deletedBy)
}

func (s *Service) GetStatusSummary(familyID models.FamilyID, profileID models.ProfileID) (*MaterialStatusSummaryResponse, error) {
	return s.repo.GetStatusSummary(familyID, profileID)
}


func (s *Service) GetEnums() (*MaterialEnumsResponse, error) {

	return &MaterialEnumsResponse{
		Types: []entities.MaterialType{
			entities.MaterialTypeMovie,
			entities.MaterialTypeShow,
		},
		WatchWith: []entities.WatchWith{
			entities.WatchWithAlone,
			entities.WatchWithPartner,
			entities.WatchWithFamily,
		},
		Statuses: []entities.Status{
			entities.StatusToWatch,
			entities.StatusInProgress,
			entities.StatusWatching,
			entities.StatusAwaitingRelease,
			entities.StatusWatched,
		},
		Priorities: []entities.Priority{
			entities.PriorityLow,
			entities.PriorityMedium,
			entities.PriorityHigh,
		},
	}, nil
}

func (s *Service) validateCreateMaterialRequest(req *CreateMaterialRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("name must be 255 characters or less")
	}

	if req.Type != "" && !s.isValidType(req.Type) {
		return fmt.Errorf("invalid type: %s", req.Type)
	}

	if req.Runtime != "" && !s.isValidRuntime(req.Runtime) {
		return fmt.Errorf("invalid runtime: %s", req.Runtime)
	}

	if len(req.SourceIDs) > 10 {
		return fmt.Errorf("too many sources (maximum 10)")
	}

	if req.WatchWith != "" && !s.isValidWatchWith(req.WatchWith) {
		return fmt.Errorf("invalid watch_with value: %s", req.WatchWith)
	}

	if req.Status != "" && !s.isValidStatus(req.Status) {
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	if req.Priority != "" && !s.isValidPriority(req.Priority) {
		return fmt.Errorf("invalid priority: %s", req.Priority)
	}

	if len(req.Notes) > 1000 {
		return fmt.Errorf("notes must be 1000 characters or less")
	}

	return nil
}

func (s *Service) validateUpdateMaterialRequest(req *UpdateMaterialRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("name must be 255 characters or less")
	}

	if !s.isValidType(req.Type) {
		return fmt.Errorf("invalid type: %s", req.Type)
	}

	if !s.isValidRuntime(req.Runtime) {
		return fmt.Errorf("invalid runtime: %s", req.Runtime)
	}

	if len(req.SourceIDs) > 10 {
		return fmt.Errorf("too many sources (maximum 10)")
	}

	if !s.isValidWatchWith(req.WatchWith) {
		return fmt.Errorf("invalid watch_with value: %s", req.WatchWith)
	}

	if !s.isValidStatus(req.Status) {
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	if !s.isValidPriority(req.Priority) {
		return fmt.Errorf("invalid priority: %s", req.Priority)
	}

	if len(req.Notes) > 1000 {
		return fmt.Errorf("notes must be 1000 characters or less")
	}

	return nil
}

func (s *Service) isValidType(t entities.MaterialType) bool {
	return t == entities.MaterialTypeMovie || t == entities.MaterialTypeShow
}


func (s *Service) isValidWatchWith(w entities.WatchWith) bool {
	return w == entities.WatchWithAlone || 
		   w == entities.WatchWithPartner || 
		   w == entities.WatchWithFamily
}

func (s *Service) isValidStatus(status entities.Status) bool {
	return status == entities.StatusToWatch ||
		   status == entities.StatusInProgress ||
		   status == entities.StatusWatching ||
		   status == entities.StatusAwaitingRelease ||
		   status == entities.StatusWatched
}

func (s *Service) isValidPriority(p entities.Priority) bool {
	return p == entities.PriorityLow ||
		   p == entities.PriorityMedium ||
		   p == entities.PriorityHigh
}

func (s *Service) isValidRuntime(r entities.Runtime) bool {
	return r == entities.RuntimeShort ||
		   r == entities.RuntimeMedium ||
		   r == entities.RuntimeLong
}