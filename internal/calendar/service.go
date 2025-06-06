package calendar

import (
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/calendar/entities"
	"github.com/chrisabs/cadence/internal/models"
	"github.com/chrisabs/cadence/internal/profile"
	"github.com/chrisabs/cadence/internal/utils/timezone"
)

const (
    MaxYearsAhead = 2
)

type Service struct {
    repo              *Repository
    profileRepo       *profile.Repository
    timezoneConverter *timezone.Converter
}

func NewService(repo *Repository, profileRepo *profile.Repository) *Service {
    return &Service{
        repo:              repo,
        profileRepo:       profileRepo,
        timezoneConverter: timezone.NewConverter(),
    }
}

func (s *Service) Create(profileID int, familyID int, req *CreateEventRequest) (*entities.Event, error) {
    profile, err := s.profileRepo.GetByID(profileID)
    if err != nil {
        return nil, fmt.Errorf("failed to get profile: %v", err)
    }
    
    startUTC, err := s.timezoneConverter.ConvertLocalToUTC(req.StartTime, profile.TimezoneName)
    if err != nil {
        return nil, fmt.Errorf("failed to convert start time: %v", err)
    }
    
    endUTC, err := s.timezoneConverter.ConvertLocalToUTC(req.EndTime, profile.TimezoneName)
    if err != nil {
        return nil, fmt.Errorf("failed to convert end time: %v", err)
    }

    event := &entities.Event{
        Title:        req.Title,
        Description:  req.Description,
        Location:     req.Location,
        StartTime:    startUTC,
        EndTime:      endUTC,
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
            
            repeatUntilUTC, err := s.timezoneConverter.ConvertLocalToUTC(*req.RepeatUntil, profile.TimezoneName)
            if err != nil {
                return nil, fmt.Errorf("failed to convert repeat until time: %v", err)
            }
            event.RecurrenceEndTime = &repeatUntilUTC  
        }
    }

    if err := s.normaliseEventTimes(event, profile.TimezoneName); err != nil {
        return nil, err
    }

    if err := s.repo.Create(event); err != nil {
        return nil, fmt.Errorf("failed to create event: %v", err)
    }

    return s.GetByID(event.ID, familyID, profileID)
}

func (s *Service) GetByID(id int, familyID int, currentProfileID int) (*entities.Event, error) {
    event, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return nil, err
    }
    
    profile, err := s.profileRepo.GetByID(currentProfileID)
    if err != nil {
        return nil, fmt.Errorf("failed to get profile: %v", err)
    }
    
    event.StartTime = s.timezoneConverter.ConvertUTCToLocal(event.StartTime, profile.TimezoneName)
    event.EndTime = s.timezoneConverter.ConvertUTCToLocal(event.EndTime, profile.TimezoneName)
    if event.RecurrenceEndTime != nil {
        convertedEnd := s.timezoneConverter.ConvertUTCToLocal(*event.RecurrenceEndTime, profile.TimezoneName)
        event.RecurrenceEndTime = &convertedEnd
    }
    
    return event, nil
}

func (s *Service) GetByDateRange(familyID int, params GetEventsParams, currentProfileID int) ([]*entities.Event, error) {
    profile, err := s.profileRepo.GetByID(currentProfileID)
    if err != nil {
        return nil, fmt.Errorf("failed to get profile: %v", err)
    }
    
    startUTC, err := s.timezoneConverter.ConvertLocalToUTC(params.StartTime, profile.TimezoneName)
    if err != nil {
        return nil, fmt.Errorf("failed to convert start time: %v", err)
    }
    
    endUTC, err := s.timezoneConverter.ConvertLocalToUTC(params.EndTime, profile.TimezoneName)
    if err != nil {
        return nil, fmt.Errorf("failed to convert end time: %v", err)
    }

    if endUTC.Before(startUTC) {
        return nil, fmt.Errorf("end time must be after start time")
    }
    
    utcParams := GetEventsParams{
        StartTime:     startUTC,
        EndTime:       endUTC,
        AssigneeIDs:   params.AssigneeIDs,
        SourceModules: params.SourceModules,
        SourceID:      params.SourceID,
    }

    events, err := s.repo.GetByDateRange(familyID, utcParams)
    if err != nil {
        return nil, fmt.Errorf("failed to get events: %v", err)
    }

    expandedEvents, err := s.expandRecurringEvents(events, startUTC, endUTC)
    if err != nil {
        return nil, fmt.Errorf("failed to expand recurring events: %v", err)
    }
    
    for _, event := range expandedEvents {
        event.StartTime = s.timezoneConverter.ConvertUTCToLocal(event.StartTime, profile.TimezoneName)
        event.EndTime = s.timezoneConverter.ConvertUTCToLocal(event.EndTime, profile.TimezoneName)
        if event.RecurrenceEndTime != nil {
            convertedEnd := s.timezoneConverter.ConvertUTCToLocal(*event.RecurrenceEndTime, profile.TimezoneName)
            event.RecurrenceEndTime = &convertedEnd
        }
    }

    return expandedEvents, nil
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
        if err := s.validateRecurringPermissions(req.UpdatedBy); err != nil {
            return nil, err
        }
        return s.updateRecurringSeries(event, familyID, req)
    }
    
    updaterProfile, err := s.profileRepo.GetByID(req.UpdatedBy)
    if err != nil {
        return nil, fmt.Errorf("failed to get updater profile: %v", err)
    }

    startUTC, err := s.timezoneConverter.ConvertLocalToUTC(req.StartTime, updaterProfile.TimezoneName)
    if err != nil {
        return nil, fmt.Errorf("failed to convert start time: %v", err)
    }
    
    endUTC, err := s.timezoneConverter.ConvertLocalToUTC(req.EndTime, updaterProfile.TimezoneName)
    if err != nil {
        return nil, fmt.Errorf("failed to convert end time: %v", err)
    }

    event.Title = req.Title
    event.Description = req.Description
    event.Location = req.Location
    event.StartTime = startUTC
    event.EndTime = endUTC
    event.AllDay = req.AllDay
    event.AssigneeID = req.AssigneeID

    if err := s.normaliseEventTimes(event, updaterProfile.TimezoneName); err != nil {
        return nil, err
    }

    if err := s.repo.Update(event); err != nil {
        return nil, fmt.Errorf("failed to update event: %v", err)
    }

    return s.GetByID(id, familyID, req.UpdatedBy)
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
        if err := s.validateRecurringPermissions(deletedBy); err != nil {
            return err
        }
    }

    return s.repo.Delete(id, familyID, deletedBy)
}

func (s *Service) RestoreDeleted(id int, familyID int) error {
    return s.repo.RestoreDeleted(id, familyID)
}

func (s *Service) updateRecurringSeries(event *entities.Event, familyID int, req *UpdateEventRequest) (*entities.Event, error) {
    updaterProfile, err := s.profileRepo.GetByID(req.UpdatedBy)
    if err != nil {
        return nil, fmt.Errorf("failed to get updater profile: %v", err)
    }

    now := time.Now().UTC()
    today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
    yesterday := today.AddDate(0, 0, -1)
    
    startUTC, err := s.timezoneConverter.ConvertLocalToUTC(req.StartTime, updaterProfile.TimezoneName)
    if err != nil {
        return nil, fmt.Errorf("failed to convert start time: %v", err)
    }
    
    endUTC, err := s.timezoneConverter.ConvertLocalToUTC(req.EndTime, updaterProfile.TimezoneName)
    if err != nil {
        return nil, fmt.Errorf("failed to convert end time: %v", err)
    }

    if event.StartTime.After(today) {
        event.Title = req.Title
        event.Description = req.Description
        event.Location = req.Location
        event.StartTime = startUTC
        event.EndTime = endUTC
        event.AllDay = req.AllDay
        event.AssigneeID = req.AssigneeID

        if err := s.normaliseEventTimes(event, updaterProfile.TimezoneName); err != nil {
            return nil, err
        }

        if err := s.repo.Update(event); err != nil {
            return nil, fmt.Errorf("failed to update future series: %v", err)
        }

        return s.GetByID(event.ID, familyID, req.UpdatedBy)
    }

    newStartTime := time.Date(
        today.Year(), today.Month(), today.Day(),
        startUTC.Hour(), startUTC.Minute(), startUTC.Second(),
        0, time.UTC,
    )
    newEndTime := time.Date(
        today.Year(), today.Month(), today.Day(),
        endUTC.Hour(), endUTC.Minute(), endUTC.Second(),
        0, time.UTC,
    )

    originalRecurrenceEnd := event.RecurrenceEndTime
    event.RecurrenceEndTime = &yesterday
    if err := s.repo.Update(event); err != nil {
        return nil, fmt.Errorf("failed to end original series: %v", err)
    }

    newSeries := &entities.Event{
        Title:             req.Title,
        Description:       req.Description,
        Location:          req.Location,
        StartTime:         newStartTime,
        EndTime:           newEndTime,
        AllDay:            req.AllDay,
        CreatedBy:         req.UpdatedBy,
        AssigneeID:        req.AssigneeID,
        FamilyID:          familyID,
        SourceModule:      event.SourceModule,
        EventType:         event.EventType,
        IsRecurring:       true,
        RecurrenceType:    event.RecurrenceType,
        RecurrenceEndTime: originalRecurrenceEnd,
    }

    if err := s.normaliseEventTimes(newSeries, updaterProfile.TimezoneName); err != nil {
        return nil, err
    }

    if err := s.repo.Create(newSeries); err != nil {
        return nil, fmt.Errorf("failed to create new series: %v", err)
    }

    return s.GetByID(newSeries.ID, familyID, req.UpdatedBy)
}

func (s *Service) UpdateRecurringInstance(req *UpdateRecurringInstanceRequest, familyID int) (*entities.Event, error) {
    event, err := s.repo.GetByID(req.EventID, familyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get event: %v", err)
    }

    if !event.IsRecurring {
        return nil, fmt.Errorf("event is not recurring")
    }

    if err := s.validateRecurringPermissions(req.UpdatedBy); err != nil {
        return nil, err
    }

    updaterProfile, err := s.profileRepo.GetByID(req.UpdatedBy)
    if err != nil {
        return nil, fmt.Errorf("failed to get updater profile: %v", err)
    }

    instanceLocal := s.timezoneConverter.ConvertUTCToLocal(req.InstanceDate, updaterProfile.TimezoneName)
    
    loc, err := time.LoadLocation(updaterProfile.TimezoneName)
    if err != nil {
        loc = time.UTC
    }
    
    exceptionDateLocal := time.Date(
        instanceLocal.Year(),
        instanceLocal.Month(),
        instanceLocal.Day(),
        0, 0, 0, 0, loc,
    )
    
    exceptionDate := exceptionDateLocal.UTC()

    existingInstance, err := s.repo.GetModifiedInstanceByDate(req.EventID, exceptionDate)
    if err != nil && err.Error() != "modified instance not found" {
        return nil, fmt.Errorf("failed to check for existing modified instance: %v", err)
    }

    if existingInstance != nil {
        if err := s.repo.Delete(existingInstance.ID, familyID, req.UpdatedBy); err != nil {
            return nil, fmt.Errorf("failed to delete existing modified instance: %v", err)
        }
    }

    updatedInstance := &entities.Event{
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
        InstanceDate:  &exceptionDate, 
        CreatedBy:     req.UpdatedBy,
    }

    if req.Title != nil {
        updatedInstance.Title = *req.Title
    }
    if req.Description != nil {
        updatedInstance.Description = req.Description
    }
    if req.Location != nil {
        updatedInstance.Location = req.Location
    }
    if req.StartTime != nil {
        startUTC, err := s.timezoneConverter.ConvertLocalToUTC(*req.StartTime, updaterProfile.TimezoneName)
        if err != nil {
            return nil, fmt.Errorf("failed to convert start time: %v", err)
        }
        updatedInstance.StartTime = startUTC
    }
    if req.EndTime != nil {
        endUTC, err := s.timezoneConverter.ConvertLocalToUTC(*req.EndTime, updaterProfile.TimezoneName)
        if err != nil {
            return nil, fmt.Errorf("failed to convert end time: %v", err)
        }
        updatedInstance.EndTime = endUTC
    }
    if req.AllDay != nil {
        updatedInstance.AllDay = *req.AllDay
    }
    if req.AssigneeID != nil {
        updatedInstance.AssigneeID = req.AssigneeID
    }

    if err := s.normaliseEventTimes(updatedInstance, updaterProfile.TimezoneName); err != nil {
        return nil, err
    }

    if err := s.repo.CreateModifiedInstance(updatedInstance); err != nil {
        return nil, fmt.Errorf("failed to create modified instance: %v", err)
    }

    if existingInstance == nil {
        if err := s.repo.CreateRecurrenceException(event.ID, exceptionDate); err != nil {
            return nil, fmt.Errorf("failed to create recurrence exception: %v", err)
        }
    }

    updatedInstance.StartTime = s.timezoneConverter.ConvertUTCToLocal(updatedInstance.StartTime, updaterProfile.TimezoneName)
    updatedInstance.EndTime = s.timezoneConverter.ConvertUTCToLocal(updatedInstance.EndTime, updaterProfile.TimezoneName)

    return updatedInstance, nil
}

func (s *Service) CancelRecurringInstance(id int, familyID int, date time.Time, cancelledBy int) error {
    event, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return fmt.Errorf("failed to get event: %v", err)
    }

    if !event.IsRecurring {
        return fmt.Errorf("event is not recurring")
    }

    if err := s.validateRecurringPermissions(cancelledBy); err != nil {
        return err
    }
    
    profile, err := s.profileRepo.GetByID(cancelledBy)
    if err != nil {
        return fmt.Errorf("failed to get profile: %v", err)
    }
    
    dateLocal := s.timezoneConverter.ConvertUTCToLocal(date, profile.TimezoneName)
    
    loc, err := time.LoadLocation(profile.TimezoneName)
    if err != nil {
        loc = time.UTC
    }
    
    exceptionDateLocal := time.Date(
        dateLocal.Year(),
        dateLocal.Month(),
        dateLocal.Day(),
        0, 0, 0, 0, loc,
    )
    
    exceptionDate := exceptionDateLocal.UTC()

    return s.repo.CreateRecurrenceException(id, exceptionDate)
}

func (s *Service) CancelFutureRecurrences(id int, familyID int, fromDate time.Time, cancelledBy int) error {
    event, err := s.repo.GetByID(id, familyID)
    if err != nil {
        return fmt.Errorf("failed to get event: %v", err)
    }

    if !event.IsRecurring {
        return fmt.Errorf("event is not recurring")
    }

    if err := s.validateRecurringPermissions(cancelledBy); err != nil {
        return err
    }

    event.RecurrenceEndTime = &fromDate
    return s.repo.Update(event)
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
    maxInstances := 500
    iterationCount := 0

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
        iterationCount++
        if iterationCount > maxInstances {
            fmt.Printf("Warning: event count exceeds: %v", maxInstances)
            break
        }
        
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
                instanceDateCopy := currentDate
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
                    InstanceDate:      &instanceDateCopy,
                    CreatedAt:         event.CreatedAt,
                    UpdatedAt:         event.UpdatedAt,
                    IsDeleted:         false,
                }
                instances = append(instances, instance)
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

func (s *Service) validateRecurringPermissions(userID int) error {
    profile, err := s.profileRepo.GetByID(userID)
    if err != nil {
        return fmt.Errorf("failed to verify role: %v", err)
    }
    if profile.Role != models.RoleParent {
        return fmt.Errorf("only parents can modify recurring events")
    }
    return nil
}

func (s *Service) normaliseEventTimes(event *entities.Event, userTimezone string) error {
    if event.AllDay {
        loc, err := time.LoadLocation(userTimezone)
        if err != nil {
            loc = time.UTC
        }
        
        startOfDayLocal := time.Date(
            event.StartTime.Year(),
            event.StartTime.Month(), 
            event.StartTime.Day(),
            0, 0, 0, 0, loc,
        )
        
        event.StartTime = startOfDayLocal.UTC()
        event.EndTime = startOfDayLocal.Add(24 * time.Hour).UTC()
    }

    if event.EndTime.Before(event.StartTime) || event.EndTime.Equal(event.StartTime) {
        return fmt.Errorf("end time must be after start time")
    }

    return nil
}