package entities

import "time"

type EventType string

const (
    EventTypeGeneral  EventType = "GENERAL"
    EventTypeChore    EventType = "CHORE"
    EventTypeMeal     EventType = "MEAL"
    EventTypeService  EventType = "SERVICE"
)

type Event struct {
    ID           int       `json:"id"`
    Title        string    `json:"title"`
    Description  *string   `json:"description,omitempty"`
    Location     *string   `json:"location,omitempty"`
    StartTime    time.Time `json:"startTime"`
    EndTime      time.Time `json:"endTime"`
    AllDay       bool      `json:"allDay"`
    CreatedBy    int       `json:"createdBy"`
    AssigneeID   int       `json:"assigneeId"`
    Type         string    `json:"type"`
    SourceModule string    `json:"sourceModule"`
    SourceID     *int      `json:"sourceId,omitempty"`
    FamilyID     int       `json:"familyId"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
    IsDeleted    bool      `json:"isDeleted"`
    DeletedAt    *time.Time `json:"deletedAt,omitempty"`
    DeletedBy    *int      `json:"deletedBy,omitempty"`
}