package calendar

import "time"

type CreateEventRequest struct {
    Title        string    `json:"title" validate:"required"`
    Description  *string   `json:"description"`
    Location     *string   `json:"location"`
    StartTime    time.Time `json:"startTime" validate:"required"`
    EndTime      time.Time `json:"endTime" validate:"required,gtfield=StartTime"`
    AllDay       bool      `json:"allDay"`
    AssigneeID   int       `json:"assigneeId" validate:"required"`
}

type UpdateEventRequest struct {
    Title        string    `json:"title" validate:"required"`
    Description  *string   `json:"description"`
    Location     *string   `json:"location"`
    StartTime    time.Time `json:"startTime" validate:"required"`
    EndTime      time.Time `json:"endTime" validate:"required,gtfield=StartTime"`
    AllDay       bool      `json:"allDay"`
    AssigneeID   int       `json:"assigneeId" validate:"required"`
}

type GetEventsParams struct {
    StartTime    time.Time `schema:"startTime" validate:"required"`
    EndTime      time.Time `schema:"endTime" validate:"required,gtfield=StartTime"`
    AssigneeID   *int     `schema:"assigneeId"`
}