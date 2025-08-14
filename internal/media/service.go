package media

import (
	"fmt"

	"github.com/chrisabs/cadence/internal/media/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateMedia(profileID models.ProfileID, familyID models.FamilyID, req *CreateMediaRequest) (*entities.Media, error) {
	if err := s.validateCreateMediaRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	if req.Status == "" {
		req.Status = entities.StatusToWatch
	}

	if req.Priority == "" {
		req.Priority = entities.PriorityMedium
	}

	return s.repo.Create(profileID, familyID, req)
}

func (s *Service) GetMediaByID(id models.MediaID, familyID models.FamilyID) (*entities.Media, error) {
	if id == "" {
		return nil, fmt.Errorf("invalid media ID")
	}

	return s.repo.GetByID(id, familyID)
}

func (s *Service) SearchMedia(familyID models.FamilyID, currentProfileID models.ProfileID, req *MediaSearchRequest) (*MediaSearchResponse, error) {
	if req.ProfileID == nil {
		req.ProfileID = &currentProfileID
	}

	media, total, err := s.repo.Search(familyID, req)
	if err != nil {
		return nil, fmt.Errorf("error searching media: %v", err)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	hasMore := offset+len(media) < total

	return &MediaSearchResponse{
		Media:   media,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}, nil
}

func (s *Service) UpdateMedia(id models.MediaID, familyID models.FamilyID, profileID models.ProfileID, req *UpdateMediaRequest) (*entities.Media, error) {
	if id == "" {
		return nil, fmt.Errorf("invalid media ID")
	}

	if err := s.validateUpdateMediaRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	return s.repo.Update(id, familyID, profileID, req)
}

func (s *Service) UpdateMediaStatus(id models.MediaID, familyID models.FamilyID, profileID models.ProfileID, status entities.Status) error {
	if id == "" {
		return fmt.Errorf("invalid media ID")
	}

	if !s.isValidStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	return s.repo.UpdateStatus(id, familyID, profileID, status)
}

func (s *Service) DeleteMedia(id models.MediaID, familyID models.FamilyID, profileID models.ProfileID, deletedBy models.ProfileID) error {
	if id == "" {
		return fmt.Errorf("invalid media ID")
	}

	return s.repo.Delete(id, familyID, profileID, deletedBy)
}

func (s *Service) GetStatusSummary(familyID models.FamilyID, profileID models.ProfileID) (*MediaStatusSummaryResponse, error) {
	return s.repo.GetStatusSummary(familyID, profileID)
}

func (s *Service) GetAllSources() ([]entities.Source, error) {
	return s.repo.GetAllSources()
}

func (s *Service) GetEnums() (*MediaEnumsResponse, error) {
	sources, err := s.repo.GetAllSources()
	if err != nil {
		return nil, fmt.Errorf("error getting sources: %v", err)
	}

	return &MediaEnumsResponse{
		Types: []entities.MediaType{
			entities.MediaTypeMovie,
			entities.MediaTypeShow,
		},
		Genres:     entities.ValidGenres,
		Sources:    sources,
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

func (s *Service) SeedSources() error {
	return s.repo.SeedSources()
}

func (s *Service) validateCreateMediaRequest(req *CreateMediaRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("name must be 255 characters or less")
	}

	if req.Type != "" && !s.isValidType(req.Type) {
		return fmt.Errorf("invalid type: %s", req.Type)
	}

	if req.Genre != "" && !s.isValidGenre(req.Genre) {
		return fmt.Errorf("invalid genre: %s", req.Genre)
	}

	if req.Runtime < 0 {
		return fmt.Errorf("runtime cannot be negative")
	}

	if req.ReleaseYear < 1800 || req.ReleaseYear > 2100 {
		return fmt.Errorf("release year must be between 1800 and 2100")
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

func (s *Service) validateUpdateMediaRequest(req *UpdateMediaRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("name must be 255 characters or less")
	}

	if !s.isValidType(req.Type) {
		return fmt.Errorf("invalid type: %s", req.Type)
	}

	if req.Genre != "" && !s.isValidGenre(req.Genre) {
		return fmt.Errorf("invalid genre: %s", req.Genre)
	}

	if req.Runtime < 0 {
		return fmt.Errorf("runtime cannot be negative")
	}

	if req.ReleaseYear < 1800 || req.ReleaseYear > 2100 {
		return fmt.Errorf("release year must be between 1800 and 2100")
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

func (s *Service) isValidType(t entities.MediaType) bool {
	return t == entities.MediaTypeMovie || t == entities.MediaTypeShow
}

func (s *Service) isValidGenre(genre string) bool {
	for _, validGenre := range entities.ValidGenres {
		if genre == validGenre {
			return true
		}
	}
	return false
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