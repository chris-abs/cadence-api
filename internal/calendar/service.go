package calendar

import (
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/calendar/entities"
	"github.com/chrisabs/cadence/internal/models"
	"github.com/chrisabs/cadence/internal/profile"
)

const (
    DefaultLimit = 50
    MaxLimit     = 200
    MaxYearsAhead = 2
)

type Service struct {
    repo        *Repository
    profileRepo *profile.Repository
}

func NewService(repo *Repository, profileRepo *profile.Repository) *Service {
    return &Service{
        repo: repo,
        profileRepo: profileRepo,
    }
}

func (s *Service) Create(familyID int, req *CreateEventRequest) (*entities.Event, error) {
    event := &entities.Event{
        Title:        req.Title,
        Description:  req.Description,
        Location:     req.Location,
        StartTime:    req.StartTime,
        EndTime:      req.EndTime,
        AllDay:       req.AllDay,
        AssigneeID:   req.AssigneeID,
        FamilyID:     familyID,
        SourceModule: "GENERAL",
        EventType:    entities.EventTypeGeneral,
    }

    if req.RepeatType != "" {
        recurrenceType := entities.RecurrenceType(req.RepeatType)
        switch recurrenceType {
        case entities.RecurrenceDaily, entities.RecurrenceWeekly, 
             entities.RecurrenceMonthly, entities.RecurrenceYearly:
            event.RecurrenceType = recurrenceType
        default:
            return nil, fmt.Errorf("invalid repeat type: %s", req.RepeatType)
        }

        if req.RepeatUntil != nil {
            maxEndTime := time.Now().AddDate(MaxYearsAhead, 0, 0)
            if req.RepeatUntil.After(maxEndTime) {
                return nil, fmt.Errorf("repeat until date cannot exceed %d years from now", MaxYearsAhead)
            }
            event.RecurrenceEndTime = req.RepeatUntil
        }

        event.IsRecurring = true
    }

    if err := s.normaliseEventTimes(event); err != nil {
        return nil, err
    }

    if err := s.repo.Create(event); err != nil {
        return nil, fmt.Errorf("failed to create event: %v", err)
    }

    return s.GetByID(event.ID, familyID)
}

func (s *Service) GetByID(id int, familyID int) (*entities.Event, error) {
    return s.repo.GetByID(id, familyID)
}

func (s *Service) GetByDateRange(familyID int, params GetEventsParams) ([]*entities.Event, bool, error) {
    if params.EndTime.Before(params.StartTime) {
        return nil, false, fmt.Errorf("end time must be after start time")
    }

    if params.Limit <= 0 {
        params.Limit = DefaultLimit
    }
    if params.Limit > MaxLimit {
        params.Limit = MaxLimit
    }

    events, total, err := s.repo.GetByDateRange(familyID, params)
    if err != nil {
        return nil, false, fmt.Errorf("failed to get events: %v", err)
    }

    hasMore := (params.Offset + len(events)) < total

    return events, hasMore, nil
}

func (s *Service) Update(id int, familyID int, req *UpdateEventRequest) (*entities.Event, error) {
    event, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get event: %v", err)
    }

    if event.SourceModule != "GENERAL" {
        return nil, fmt.Errorf("cannot update events from %s module directly", event.SourceModule)
    }

    if event.IsRecurring {
        profile, err := s.profileRepo.GetByID(req.UpdatedBy)
        if err != nil {
            return nil, fmt.Errorf("failed to verify role: %v", err)
        }
        if profile.Role != models.RoleParent {
            return nil, fmt.Errorf("only parents can modify recurring events")
        }
    }

    event.Title = req.Title
    event.Description = req.Description
    event.Location = req.Location
    event.StartTime = req.StartTime
    event.EndTime = req.EndTime
    event.AllDay = req.AllDay
    event.AssigneeID = req.AssigneeID

    if err := s.normaliseEventTimes(event); err != nil {
        return nil, err
    }

    if err := s.repo.Update(event); err != nil {
        return nil, fmt.Errorf("failed to update event: %v", err)
    }

    return s.GetByID(id, familyID)
}

func (s *Service) ModifyRecurringInstance(req *ModifyRecurringInstanceRequest, familyID int) (*entities.Event, error) {
    originalEvent, err := s.repo.GetByID(req.EventID, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get event: %v", err)
    }

    if !originalEvent.IsRecurring {
        return nil, fmt.Errorf("event is not recurring")
    }

    profile, err := s.profileRepo.GetByID(req.UpdatedBy)
    if err != nil {
        return nil, fmt.Errorf("failed to verify role: %v", err)
    }
    if profile.Role != models.RoleParent {
        return nil, fmt.Errorf("only parents can modify recurring events")
    }

    modifiedInstance := &entities.Event{
        Title:         originalEvent.Title,
        Description:   originalEvent.Description,
        Location:      originalEvent.Location,
        StartTime:     originalEvent.StartTime,
        EndTime:       originalEvent.EndTime,
        AllDay:        originalEvent.AllDay,
        AssigneeID:    originalEvent.AssigneeID,
        FamilyID:      familyID,
        SourceModule:  originalEvent.SourceModule,
        EventType:     originalEvent.EventType,
        IsException:   true,
        ParentEventID: &originalEvent.ID,
    }

    if req.Title != nil {
        modifiedInstance.Title = *req.Title
    }
    if req.Description != nil {
        modifiedInstance.Description = req.Description
    }
    if req.Location != nil {
        modifiedInstance.Location = req.Location
    }
    if req.StartTime != nil {
        modifiedInstance.StartTime = *req.StartTime
    }
    if req.EndTime != nil {
        modifiedInstance.EndTime = *req.EndTime
    }
    if req.AllDay != nil {
        modifiedInstance.AllDay = *req.AllDay
    }
    if req.AssigneeID != nil {
        modifiedInstance.AssigneeID = *req.AssigneeID
    }

    if err := s.normaliseEventTimes(modifiedInstance); err != nil {
        return nil, err
    }

    if err := s.repo.CreateModifiedInstance(modifiedInstance); err != nil {
        return nil, fmt.Errorf("failed to create modified instance: %v", err)
    }

    return modifiedInstance, nil
}

func (s *Service) Delete(id int, familyID int, deletedBy int) error {
    event, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return fmt.Errorf("failed to get event: %v", err)
    }

    if event.SourceModule != "GENERAL" {
        return fmt.Errorf("cannot delete events from %s module directly", event.SourceModule)
    }

    if event.IsRecurring {
        profile, err := s.profileRepo.GetByID(deletedBy)
        if err != nil {
            return fmt.Errorf("failed to verify role: %v", err)
        }
        if profile.Role != models.RoleParent {
            return fmt.Errorf("only parents can delete recurring events")
        }
    }

    return s.repo.Delete(id, familyID, deletedBy)
}

func (s *Service) RestoreDeleted(id int, familyID int) error {
    return s.repo.RestoreDeleted(id, familyID)
}

func (s *Service) CancelRecurringInstance(id int, familyID int, date time.Time, cancelledBy int) error {
    event, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return fmt.Errorf("failed to get event: %v", err)
    }

    if !event.IsRecurring {
        return fmt.Errorf("event is not recurring")
    }

    profile, err := s.profileRepo.GetByID(cancelledBy)
    if err != nil {
        return fmt.Errorf("failed to verify role: %v", err)
    }
    if profile.Role != models.RoleParent {
        return fmt.Errorf("only parents can cancel recurring events")
    }

    return s.repo.CreateRecurrenceException(id, date)
}

func (s *Service) CancelFutureRecurrences(id int, familyID int, fromDate time.Time, cancelledBy int) error {
    event, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return fmt.Errorf("failed to get event: %v", err)
    }

    if !event.IsRecurring {
        return fmt.Errorf("event is not recurring")
    }

    profile, err := s.profileRepo.GetByID(cancelledBy)
    if err != nil {
        return fmt.Errorf("failed to verify role: %v", err)
    }
    if profile.Role != models.RoleParent {
        return fmt.Errorf("only parents can cancel recurring events")
    }

    event.RecurrenceEndTime = &fromDate
    return s.repo.Update(event)
}

func (s *Service) normaliseEventTimes(event *entities.Event) error {
    if event.EndTime.Before(event.StartTime) {
        return fmt.Errorf("end time must be after start time")
    }

    if event.AllDay {
        event.StartTime = time.Date(
            event.StartTime.Year(),
            event.StartTime.Month(),
            event.StartTime.Day(),
            0, 0, 0, 0,
            time.UTC,
        )
        
        event.EndTime = time.Date(
            event.StartTime.Year(),  
            event.StartTime.Month(),
            event.StartTime.Day(),
            23, 59, 59, 999999999,
            time.UTC,
        )
    }

    return nil
}