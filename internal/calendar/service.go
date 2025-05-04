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

func (s *Service) Create(createdBy int, familyID int, req *CreateEventRequest) (*entities.Event, error) {
    if err := s.validateEventTimes(&req.StartTime, &req.EndTime, req.AllDay); err != nil {
        return nil, err
    }

    event := &entities.Event{
        Title:        req.Title,
        Description:  req.Description,
        Location:     req.Location,
        StartTime:    req.StartTime,
        EndTime:      req.EndTime,
        AllDay:       req.AllDay,
        CreatedBy:    createdBy,
        AssigneeID:   req.AssigneeID,
        FamilyID:     familyID,
        Type:         string(entities.EventTypeGeneral),
        SourceModule: string(entities.EventTypeGeneral),
    }

    if err := s.repo.Create(event); err != nil {
        return nil, fmt.Errorf("failed to create event: %v", err)
    }

    return s.repo.GetByID(event.ID, familyID)
}

func (s *Service) Update(id int, familyID int, req *UpdateEventRequest) (*entities.Event, error) {
    if err := s.validateEventTimes(&req.StartTime, &req.EndTime, req.AllDay); err != nil {
        return nil, err
    }

    event, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get event: %v", err)
    }

    
    if event.SourceModule != string(entities.EventTypeGeneral) {
        return nil, fmt.Errorf("cannot update events from %s module directly", event.SourceModule)
    }

    event.Title = req.Title
    event.Description = req.Description
    event.Location = req.Location
    event.StartTime = req.StartTime
    event.EndTime = req.EndTime
    event.AllDay = req.AllDay
    event.AssigneeID = req.AssigneeID

    if err := s.repo.Update(event); err != nil {
        return nil, fmt.Errorf("failed to update event: %v", err)
    }

    return s.repo.GetByID(id, familyID)
}

func (s *Service) GetByID(id int, familyID int) (*entities.Event, error) {
    return s.repo.GetByID(id, familyID)
}

func (s *Service) GetByDateRange(familyID int, params GetEventsParams) ([]*entities.Event, error) {
    if err := s.validateEventTimes(&params.StartTime, &params.EndTime, false); err != nil {
        return nil, err
    }

    return s.repo.GetByDateRange(familyID, params)
}

func (s *Service) Delete(id int, familyID int, deletedBy int) error {
    event, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return fmt.Errorf("failed to get event: %v", err)
    }

    
    if event.SourceModule != string(entities.EventTypeGeneral) {
        return fmt.Errorf("cannot delete events from %s module directly", event.SourceModule)
    }

    return s.repo.Delete(id, familyID, deletedBy)
}

func (s *Service) RestoreDeleted(id int, familyID int) error {
    return s.repo.RestoreDeleted(id, familyID)
}

func (s *Service) validateEventTimes(start, end *time.Time, allDay bool) error {
    if start.After(*end) {
        return fmt.Errorf("end time must be after start time")
    }

    if allDay {
        
        *start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
        
        *end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, time.UTC)
    }

    return nil
}