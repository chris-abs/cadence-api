package calendar

import (
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/calendar/entities"
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

func (s *Service) GetByID(id int, familyID int) (*entities.Event, error) {
    return s.repo.GetByID(id, familyID)
}

func (s *Service) Create(createdBy int, familyID int, req *CreateEventRequest) (*entities.Event, error) {
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
        SourceModule: "GENERAL",
        EventType:    entities.EventTypeGeneral,
    }

    // Handle recurring events
    if req.RepeatType != "" {
        // Verify parent role for recurring events
        role, err := s.profileRepo.GetRole(createdBy)
        if err != nil {
            return nil, fmt.Errorf("failed to verify role: %v", err)
        }
        if role != "PARENT" {
            return nil, fmt.Errorf("only parents can create recurring events")
        }

        // Validate repeat type
        recurrenceType := entities.RecurrenceType(req.RepeatType)
        switch recurrenceType {
        case entities.RecurrenceDaily, entities.RecurrenceWeekly, 
             entities.RecurrenceMonthly, entities.RecurrenceYearly:
            event.RecurrenceType = recurrenceType
        default:
            return nil, fmt.Errorf("invalid repeat type: %s", req.RepeatType)
        }

        // Validate end date
        if req.RepeatUntil != nil {
            maxEndTime := time.Now().AddDate(MaxYearsAhead, 0, 0)
            if req.RepeatUntil.After(maxEndTime) {
                return nil, fmt.Errorf("repeat until date cannot exceed %d years from now", MaxYearsAhead)
            }
            event.RecurrenceEndTime = req.RepeatUntil
        }

        event.IsRecurring = true
        event.RecurrenceState = entities.RecurrenceActive
    }

    if err := s.normaliseEventTimes(event); err != nil {
        return nil, err
    }

    if err := s.repo.Create(event); err != nil {
        return nil, fmt.Errorf("failed to create event: %v", err)
    }

    return s.GetByID(event.ID, familyID)
}

func (s *Service) GetByDateRange(familyID int, params GetEventsParams) ([]*entities.Event, bool, error) {
    if params.EndTime.Before(params.StartTime) {
        return nil, false, fmt.Errorf("end time must be after start time")
    }

    // Handle pagination
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

    // Check parent role for recurring events
    if event.IsRecurring {
        role, err := s.profileRepo.GetRole(req.UpdatedBy)
        if err != nil {
            return nil, fmt.Errorf("failed to verify role: %v", err)
        }
        if role != "PARENT" {
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

func (s *Service) Delete(id int, familyID int, deletedBy int) error {
    event, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return fmt.Errorf("failed to get event: %v", err)
    }

    if event.SourceModule != "GENERAL" {
        return fmt.Errorf("cannot delete events from %s module directly", event.SourceModule)
    }

    // Check parent role for recurring events
    if event.IsRecurring {
        role, err := s.profileRepo.GetRole(deletedBy)
        if err != nil {
            return fmt.Errorf("failed to verify role: %v", err)
        }
        if role != "PARENT" {
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

    role, err := s.profileRepo.GetRole(cancelledBy)
    if err != nil {
        return fmt.Errorf("failed to verify role: %v", err)
    }
    if role != "PARENT" {
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

    role, err := s.profileRepo.GetRole(cancelledBy)
    if err != nil {
        return fmt.Errorf("failed to verify role: %v", err)
    }
    if role != "PARENT" {
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
        // Set to start of day in UTC
        event.StartTime = time.Date(
            event.StartTime.Year(),
            event.StartTime.Month(),
            event.StartTime.Day(),
            0, 0, 0, 0,
            time.UTC,
        )
        
        // Set to end of the same day in UTC
        event.EndTime = time.Date(
            event.StartTime.Year(),  // Use StartTime to ensure same day
            event.StartTime.Month(),
            event.StartTime.Day(),
            23, 59, 59, 999999999,
            time.UTC,
        )
    }

    return nil
}