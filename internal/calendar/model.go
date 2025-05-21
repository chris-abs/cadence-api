package calendar

import (
	"time"

	"github.com/chrisabs/cadence/internal/calendar/entities"
)

type GetEventsParams struct {
    StartTime    time.Time `schema:"startTime"`
    EndTime      time.Time `schema:"endTime"`
    AssigneeIDs  []int     `schema:"assigneeIds,omitempty"` 
    ModuleIDs    []string  `schema:"moduleIds,omitempty"`
    SourceID     *int      `schema:"sourceId,omitempty"`
}

type CreateEventRequest struct {
    Title       string    `json:"title" validate:"required"`
    Description *string   `json:"description"`
    Location    *string   `json:"location"`
    StartTime   time.Time `json:"startTime" validate:"required"`
    EndTime     time.Time `json:"endTime" validate:"required,gtfield=StartTime"`
    AllDay      bool      `json:"allDay"`
    AssigneeID  int       `json:"assigneeId" validate:"required"`
    RepeatType   string    `json:"repeatType,omitempty"`
    RepeatUntil  *time.Time `json:"repeatUntil,omitempty"`
}

type UpdateEventRequest struct {
    Title       string    `json:"title" validate:"required"`
    Description *string   `json:"description"`
    Location    *string   `json:"location"`
    StartTime   time.Time `json:"startTime" validate:"required"`
    EndTime     time.Time `json:"endTime" validate:"required,gtfield=StartTime"`
    AllDay      bool      `json:"allDay"`
    AssigneeID  int       `json:"assigneeId" validate:"required"`
    UpdatedBy   int       `json:"updatedBy" validate:"required"`
}

type ModifyRecurringInstanceRequest struct {
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