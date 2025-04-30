package calendar

import (
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/calendar/entities"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateEvent(familyID int, profileID int, req *CreateEventRequest) (*entities.Event, error) {
    if req.Title == "" {
        return nil, fmt.Errorf("event title is required")
    }

    if req.Start.After(req.End) {
        return nil, fmt.Errorf("event start time must be before end time")
    }

    now := time.Now().UTC()
    
    event := &entities.Event{
        Title:       req.Title,
        Description: req.Description,
        Start:      req.Start,
        End:        req.End,
        AllDay:     req.AllDay,
        CreatedBy:  profileID,
        AssigneeIDs: req.AssigneeIDs,
        Color:      req.Color,
        Type:       "GENERAL",
        FamilyID:   familyID,
        CreatedAt:  now,
        UpdatedAt:  now,
    }

    if err := s.repo.Create(event); err != nil {
        return nil, fmt.Errorf("failed to create event: %v", err)
    }

    return s.GetEventByID(event.ID, familyID)
}

func (s *Service) GetEventByID(id int, familyID int) (*entities.Event, error) {
    return s.repo.GetByID(id, familyID)
}

func (s *Service) GetEvents(familyID int, params GetEventsParams) ([]*entities.Event, error) {
    if params.Start.After(params.End) {
        return nil, fmt.Errorf("start date must be before end date")
    }

    return s.repo.GetByDateRange(familyID, params)
}

func (s *Service) UpdateEvent(id int, familyID int, profileID int, req *UpdateEventRequest) (*entities.Event, error) {
    existing, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, err
    }

    if existing.Type != "GENERAL" {
        return nil, fmt.Errorf("cannot update non-general events directly")
    }

    if req.Start.After(req.End) {
        return nil, fmt.Errorf("event start time must be before end time")
    }

    event := &entities.Event{
        ID:          id,
        Title:       req.Title,
        Description: req.Description,
        Start:       req.Start,
        End:         req.End,
        AllDay:      req.AllDay,
        AssigneeIDs: req.AssigneeIDs,
        Color:       req.Color,
        FamilyID:    familyID,
        UpdatedAt:   time.Now().UTC(),
    }

    if err := s.repo.Update(event); err != nil {
        return nil, fmt.Errorf("failed to update event: %v", err)
    }

    return s.repo.GetByID(id, familyID)
}

func (s *Service) DeleteEvent(id int, familyID int, deletedBy int) error {
    existing, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return err
    }

    if existing.Type != "GENERAL" {
        return fmt.Errorf("cannot delete non-general events directly")
    }

    return s.repo.Delete(id, familyID, deletedBy)
}

func (s *Service) RestoreEvent(id int, familyID int) error {
    if err := s.repo.RestoreDeleted(id, familyID); err != nil {
        return fmt.Errorf("failed to restore event: %v", err)
    }
    return nil
}