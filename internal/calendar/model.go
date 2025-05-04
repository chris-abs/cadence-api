package calendar

import "time"

type CreateEventRequest struct {
    Title       string    `json:"title"`
    Description string    `json:"description,omitempty"`
    Location    string    `json:"location,omitempty"`
    Start       time.Time `json:"start"`
    End         time.Time `json:"end"`
    AllDay      bool      `json:"allDay"`
    AssigneeIDs []int     `json:"assigneeIds"`
    Color       *string   `json:"color,omitempty"`
}

type UpdateEventRequest struct {
    Title       string    `json:"title"`
    Description string    `json:"description,omitempty"`
    Location    string     `json:"location,omitempty"`
    Start       time.Time `json:"start"`
    End         time.Time `json:"end"`
    AllDay      bool      `json:"allDay"`
    AssigneeIDs []int     `json:"assigneeIds"`
    Color       *string   `json:"color,omitempty"`
}

type GetEventsParams struct {
    Start       time.Time `json:"start"`
    End         time.Time `json:"end"`
    AssigneeIDs []int     `json:"assigneeIds,omitempty"`
    Types       []string  `json:"types,omitempty"`
    ModuleIDs   []string  `json:"moduleIds,omitempty"`
}