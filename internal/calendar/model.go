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
    Limit        int       `schema:"limit,omitempty"`
    Offset       int       `schema:"offset,omitempty"`
}

type CreateEventRequest struct {
    Title       string     `json:"title" validate:"required"`
    Description *string    `json:"description"`
    Location    *string    `json:"location"`
    StartTime   time.Time  `json:"startTime" validate:"required"`
    EndTime     time.Time  `json:"endTime" validate:"required,gtfield=StartTime"`
    AllDay      bool       `json:"allDay"`
    AssigneeID  int        `json:"assigneeId" validate:"required"`
    RepeatType  string     `json:"repeatType,omitempty"` 
    RepeatUntil *time.Time `json:"repeatUntil,omitempty"`
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

type EventResponse struct {
    *entities.Event
    HasMore bool `json:"hasMore,omitempty"`
}