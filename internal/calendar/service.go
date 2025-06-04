package calendar

import (
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/calendar/entities"
	"github.com/chrisabs/cadence/internal/models"
	"github.com/chrisabs/cadence/internal/profile"
)

const (
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

func (s *Service) Create(profileID int, familyID int, req *CreateEventRequest) (*entities.Event, error) {
    event := &entities.Event{
        Title:        req.Title,
        Description:  req.Description,
        Location:     req.Location,
        StartTime:    req.StartTime,
        EndTime:      req.EndTime,
        AllDay:       req.AllDay,
        CreatedBy:    profileID,
        AssigneeID:   req.AssigneeID,
        FamilyID:     familyID,
        SourceModule: "GENERAL",
        EventType:    entities.EventTypeGeneral,
        IsRecurring:  false,
        RecurrenceType: nil, 
    }

    if req.RepeatType != nil && *req.RepeatType != "" {
        recurrenceType := entities.RecurrenceType(*req.RepeatType)
        switch recurrenceType {
        case entities.RecurrenceDaily, entities.RecurrenceWeekly, 
             entities.RecurrenceMonthly, entities.RecurrenceYearly:
            event.RecurrenceType = &recurrenceType 
            event.IsRecurring = true
        default:
            return nil, fmt.Errorf("invalid repeat type: %s", *req.RepeatType)
        }

        if req.RepeatUntil != nil {
            maxEndTime := time.Now().AddDate(MaxYearsAhead, 0, 0)
            if req.RepeatUntil.After(maxEndTime) {
                return nil, fmt.Errorf("repeat until date cannot exceed %d years from now", MaxYearsAhead)
            }
            event.RecurrenceEndTime = req.RepeatUntil
        }
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

func (s *Service) GetByDateRange(familyID int, params GetEventsParams) ([]*entities.Event, error) {
    fmt.Printf("DEBUG: GetByDateRange called for family %d, start: %v, end: %v\n", familyID, params.StartTime, params.EndTime)
    
    if params.EndTime.Before(params.StartTime) {
        return nil, fmt.Errorf("end time must be after start time")
    }

    events, err := s.repo.GetByDateRange(familyID, params)
    if err != nil {
        return nil, fmt.Errorf("failed to get events: %v", err)
    }

    fmt.Printf("DEBUG: Found %d events from database\n", len(events))
    for _, event := range events {
        fmt.Printf("DEBUG: Event: %s (ID: %d, IsRecurring: %t, IsException: %t)\n", 
            event.Title, event.ID, event.IsRecurring, event.IsException)
    }

    expandedEvents, err := s.expandRecurringEvents(events, params.StartTime, params.EndTime)
    if err != nil {
        return nil, fmt.Errorf("failed to expand recurring events: %v", err)
    }

    for _, event := range expandedEvents {
        if event.Title == "shopping" && event.InstanceDate != nil {
            fmt.Printf("DEBUG: Returning shopping event - ID: %d, startTime: %s, instanceDate: %s\n", 
                event.ID, event.StartTime.Format("2006-01-02T15:04:05Z"), 
                event.InstanceDate.Format("2006-01-02T15:04:05Z"))
        }
    }

    fmt.Printf("DEBUG: After expansion, returning %d events\n", len(expandedEvents))
    return expandedEvents, nil
}

func (s *Service) expandRecurringEvents(events []*entities.Event, startTime, endTime time.Time) ([]*entities.Event, error) {
    var result []*entities.Event
    var recurringEvents []*entities.Event
    var recurringEventIDs []int

    for _, event := range events {
        if event.IsRecurring && !event.IsException {
            recurringEvents = append(recurringEvents, event)
            recurringEventIDs = append(recurringEventIDs, event.ID)
        } else {
            result = append(result, event)
        }
    }

    if len(recurringEvents) == 0 {
        return result, nil
    }

    exceptions, err := s.repo.GetExceptionsForEvents(recurringEventIDs)
    if err != nil {
        return nil, fmt.Errorf("failed to get exceptions: %v", err)
    }

    modifiedInstances, err := s.repo.GetModifiedInstancesInDateRange(events[0].FamilyID, startTime, endTime, recurringEventIDs)
    if err != nil {
        return nil, fmt.Errorf("failed to get modified instances: %v", err)
    }

    modifiedInstanceMap := make(map[string]*entities.Event)
    for _, instance := range modifiedInstances {
        if instance.ParentEventID != nil && instance.InstanceDate != nil {
            key := fmt.Sprintf("%d-%s", *instance.ParentEventID, instance.InstanceDate.Format("2006-01-02"))
            modifiedInstanceMap[key] = instance
        }
    }

    for _, recurringEvent := range recurringEvents {
        instances := s.generateRecurringInstances(recurringEvent, startTime, endTime, exceptions[recurringEvent.ID], modifiedInstanceMap)
        result = append(result, instances...)
    }

    return result, nil
}

func (s *Service) generateRecurringInstances(event *entities.Event, startTime, endTime time.Time, cancelledDates []time.Time, modifiedInstanceMap map[string]*entities.Event) []*entities.Event {
    var instances []*entities.Event

    cancelled := make(map[string]bool)
    for _, date := range cancelledDates {
        dateKey := date.Format("2006-01-02")
        cancelled[dateKey] = true
    }

    duration := event.EndTime.Sub(event.StartTime)

    currentDate := event.StartTime
    recurrenceEnd := endTime
    if event.RecurrenceEndTime != nil && event.RecurrenceEndTime.Before(endTime) {
        recurrenceEnd = *event.RecurrenceEndTime
    }

    for currentDate.Before(recurrenceEnd) {
        occurrenceEnd := currentDate.Add(duration)
        if occurrenceEnd.After(startTime) && currentDate.Before(endTime) {
            instanceDateKey := currentDate.Format("2006-01-02")
            
            if cancelled[instanceDateKey] {
                currentDate = s.getNextOccurrence(currentDate, *event.RecurrenceType)
                continue
            }

            modifiedKey := fmt.Sprintf("%d-%s", event.ID, instanceDateKey)
            if modifiedInstance, exists := modifiedInstanceMap[modifiedKey]; exists {
                instances = append(instances, modifiedInstance)
            } else {
                instanceDate := currentDate
                
                instance := &entities.Event{
                    ID:                event.ID, 
                    Title:             event.Title,
                    Description:       event.Description,
                    Location:          event.Location,
                    StartTime:         currentDate,
                    EndTime:           currentDate.Add(duration),
                    AllDay:            event.AllDay,
                    CreatedBy:         event.CreatedBy,
                    AssigneeID:        event.AssigneeID,
                    Assignee:          event.Assignee,
                    SourceModule:      event.SourceModule,
                    SourceID:          event.SourceID,
                    FamilyID:          event.FamilyID,
                    EventType:         event.EventType,
                    IsRecurring:       true,
                    RecurrenceType:    event.RecurrenceType,
                    RecurrenceEndTime: event.RecurrenceEndTime,
                    IsException:       false,
                    ParentEventID:     nil, 
                    InstanceDate:      &instanceDate,
                    CreatedAt:         event.CreatedAt,
                    UpdatedAt:         event.UpdatedAt,
                    IsDeleted:         false,
                }

                if err := s.normaliseEventTimes(instance); err == nil {
                    instances = append(instances, instance)
                }
            }
        }

        currentDate = s.getNextOccurrence(currentDate, *event.RecurrenceType)
        
        if currentDate.After(time.Now().AddDate(MaxYearsAhead, 0, 0)) {
            break
        }
    }

    return instances
}

func (s *Service) getNextOccurrence(current time.Time, recurrenceType entities.RecurrenceType) time.Time {
    switch recurrenceType {
    case entities.RecurrenceDaily:
        return current.AddDate(0, 0, 1)
    case entities.RecurrenceWeekly:
        return current.AddDate(0, 0, 7)
    case entities.RecurrenceMonthly:
        return current.AddDate(0, 1, 0)
    case entities.RecurrenceYearly:
        return current.AddDate(1, 0, 0)
    default:
        return current.AddDate(0, 0, 1) 
    }
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
    event, err := s.repo.GetByID(req.EventID, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get event: %v", err)
    }

    if !event.IsRecurring {
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
        Title:         event.Title,
        Description:   event.Description,
        Location:      event.Location,
        StartTime:     event.StartTime,
        EndTime:       event.EndTime,
        AllDay:        event.AllDay,
        AssigneeID:    event.AssigneeID,
        FamilyID:      familyID,
        SourceModule:  event.SourceModule,
        EventType:     event.EventType,
        IsException:   true,
        ParentEventID: &event.ID,
        InstanceDate:  &req.InstanceDate, 
        CreatedBy:     req.UpdatedBy,
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
        modifiedInstance.AssigneeID = req.AssigneeID
    }

    if err := s.normaliseEventTimes(modifiedInstance); err != nil {
        return nil, err
    }

    if err := s.repo.CreateModifiedInstance(modifiedInstance); err != nil {
        return nil, fmt.Errorf("failed to create modified instance: %v", err)
    }

    exceptionDate := time.Date(
        req.InstanceDate.Year(),
        req.InstanceDate.Month(),
        req.InstanceDate.Day(),
        0, 0, 0, 0, time.UTC,
    )
    
    fmt.Printf("Creating exception for eventID: %d, date: %v\n", event.ID, exceptionDate)
    
    if err := s.repo.CreateRecurrenceException(event.ID, exceptionDate); err != nil {
        return nil, fmt.Errorf("failed to create recurrence exception: %v", err)
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
    if event.AllDay {
        startOfDay := time.Date(
            event.StartTime.Year(),
            event.StartTime.Month(),
            event.StartTime.Day(),
            0, 0, 0, 0,
            time.UTC,
        )
        
        event.StartTime = startOfDay
        event.EndTime = startOfDay.Add(24 * time.Hour) 
    }

    if event.EndTime.Before(event.StartTime) || event.EndTime.Equal(event.StartTime) {
        return fmt.Errorf("end time must be after start time")
    }

    return nil
}