package calendar

import (
	"time"

	"github.com/chrisabs/cadence/internal/calendar/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type GetEventsParams struct {
    StartTime     time.Time `schema:"startTime"`
    EndTime       time.Time `schema:"endTime"`
    AssigneeIDs   []int     `schema:"assigneeIds,omitempty"` 
    SourceModules []string  `schema:"sourceModules,omitempty"`
    SourceID      *int      `schema:"sourceId,omitempty"`
}

type CreateEventRequest struct {
    Title       string        `json:"title" validate:"required"`
    Description *string       `json:"description"`
    Location    *string       `json:"location"`
    StartTime   time.Time     `json:"startTime" validate:"required"`
    EndTime     time.Time     `json:"endTime" validate:"required,gtfield=StartTime"`
    AllDay      bool          `json:"allDay"`
    AssigneeID  *int          `json:"assigneeId"`
    RepeatType  *string       `json:"repeatType,omitempty"`
    RepeatUntil  *time.Time   `json:"repeatUntil,omitempty"`
}

type UpdateEventRequest struct {
    Title       string    `json:"title" validate:"required"`
    Description *string   `json:"description"`
    Location    *string   `json:"location"`
    StartTime   time.Time `json:"startTime" validate:"required"`
    EndTime     time.Time `json:"endTime" validate:"required,gtfield=StartTime"`
    AllDay      bool      `json:"allDay"`
    AssigneeID  *int      `json:"assigneeId"`
    UpdatedBy   int       `json:"updatedBy" validate:"required"`
}

type UpdateRecurringInstanceRequest struct {
    EventID     int        `json:"eventId" validate:"required"`
    InstanceDate time.Time `json:"instanceDate" validate:"required"`
    UpdatedBy   int        `json:"updatedBy" validate:"required"`
    Title       *string    `json:"title,omitempty"`
    Description *string    `json:"description,omitempty"`
    Location    *string    `json:"location,omitempty"`
    StartTime   *time.Time `json:"startTime,omitempty"`
    EndTime     *time.Time `json:"endTime,omitempty"`
    AllDay      *bool      `json:"allDay,omitempty"`
    AssigneeID  *int       `json:"assigneeId,omitempty"`
}

type PaginatedEvents struct {
    Events     []*entities.Event `json:"events"`
    HasMore    bool              `json:"hasMore"`
}

type CalendarResponse struct {
    Events   []*CalendarEventDTO    `json:"events"`
    Profiles map[int]*models.Profile `json:"profiles"`
}

type CalendarEventDTO struct {
    ID                int                       `json:"id"`
    Title             string                    `json:"title"`
    Description       *string                   `json:"description,omitempty"`
    Location          *string                   `json:"location,omitempty"`
    StartTime         time.Time                 `json:"startTime"`
    EndTime           time.Time                 `json:"endTime"`
    AllDay            bool                      `json:"allDay"`
    AssigneeID        *int                      `json:"assigneeId,omitempty"`
    SourceModule      string                    `json:"sourceModule"`
    SourceID          *int                      `json:"sourceId,omitempty"`
    FamilyID          int                       `json:"familyId"`
    IsRecurring       bool                      `json:"isRecurring"`
    RecurrenceType    *entities.RecurrenceType  `json:"recurrenceType,omitempty"`
    RecurrenceEndTime *time.Time                `json:"recurrenceEndTime,omitempty"`
    IsException       bool                      `json:"isException"`
    ParentEventID     *int                      `json:"parentEventId,omitempty"`
    InstanceDate      *time.Time                `json:"instanceDate,omitempty"`
    CreatedAt         time.Time                 `json:"createdAt"`
    UpdatedAt         time.Time                 `json:"updatedAt"`
    IsDeleted         bool                      `json:"isDeleted"`
    DeletedAt         *time.Time                `json:"deletedAt,omitempty"`
    DeletedBy         *int                      `json:"deletedBy,omitempty"`
}

func NormalizeEventsResponse(events []*entities.Event) *CalendarResponse {
    profiles := make(map[int]*models.Profile)
    eventDTOs := make([]*CalendarEventDTO, len(events))
    
    for i, event := range events {
        if event.Assignee != nil {
            profiles[event.Assignee.ID] = event.Assignee
        }
        
        eventDTOs[i] = &CalendarEventDTO{
            ID:                event.ID,
            Title:             event.Title,
            Description:       event.Description,
            Location:          event.Location,
            StartTime:         event.StartTime,
            EndTime:           event.EndTime,
            AllDay:            event.AllDay,
            AssigneeID:        event.AssigneeID,
            SourceModule:      event.SourceModule,
            SourceID:          event.SourceID,
            FamilyID:          event.FamilyID,
            IsRecurring:       event.IsRecurring,
            RecurrenceType:    event.RecurrenceType,
            RecurrenceEndTime: event.RecurrenceEndTime,
            IsException:       event.IsException,
            ParentEventID:     event.ParentEventID,
            InstanceDate:      event.InstanceDate,
            CreatedAt:         event.CreatedAt,
            UpdatedAt:         event.UpdatedAt,
            IsDeleted:         event.IsDeleted,
            DeletedAt:         event.DeletedAt,
            DeletedBy:         event.DeletedBy,
        }
    }
    
    return &CalendarResponse{
        Events:   eventDTOs,
        Profiles: profiles,
    }
}