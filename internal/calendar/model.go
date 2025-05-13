package calendar

import "time"

type GetEventsParams struct {
    StartTime    time.Time `schema:"startTime"`
    EndTime      time.Time `schema:"endTime"`
    AssigneeID   *int      `schema:"assigneeId,omitempty"`
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
}

type UpdateEventRequest struct {
    Title       string    `json:"title" validate:"required"`
    Description *string   `json:"description"`
    Location    *string   `json:"location"`
    StartTime   time.Time `json:"startTime" validate:"required"`
    EndTime     time.Time `json:"endTime" validate:"required,gtfield=StartTime"`
    AllDay      bool      `json:"allDay"`
    AssigneeID  int       `json:"assigneeId" validate:"required"`
}