package calendar

import datetime "github.com/chrisabs/cadence/internal/types"

type CreateEventRequest struct {
    Title       string           `json:"title"`
    Description string           `json:"description"`
    Start       datetime.DateTime `json:"start"`
    End         datetime.DateTime `json:"end"`
    AllDay      bool             `json:"allDay"`
    AssigneeIDs []int            `json:"assigneeIds"`
    Color       *string          `json:"color,omitempty"`
}

type UpdateEventRequest struct {
    Title       string           `json:"title"`
    Description string           `json:"description"`
    Start       datetime.DateTime `json:"start"`
    End         datetime.DateTime `json:"end"`
    AllDay      bool             `json:"allDay"`
    AssigneeIDs []int            `json:"assigneeIds"`
    Color       *string          `json:"color,omitempty"`
}

type GetEventsParams struct {
    Start       datetime.DateTime `json:"start"`
    End         datetime.DateTime `json:"end"`
    AssigneeIDs []int            `json:"assigneeIds,omitempty"`
    Types       []string         `json:"types,omitempty"`
    ModuleIDs   []string         `json:"moduleIds,omitempty"`
}