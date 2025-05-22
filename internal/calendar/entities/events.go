package entities

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type EventType string
type RecurrenceType string

const (
    EventTypeGeneral  EventType = "GENERAL"
    EventTypeChore    EventType = "CHORE"
    EventTypeMeal     EventType = "MEAL"
    EventTypeService  EventType = "SERVICE"
)

const (
    RecurrenceNone    RecurrenceType = ""
    RecurrenceDaily   RecurrenceType = "DAILY"
    RecurrenceWeekly  RecurrenceType = "WEEKLY"
    RecurrenceMonthly RecurrenceType = "MONTHLY"
    RecurrenceYearly  RecurrenceType = "YEARLY"
)

type Event struct {
    ID           int             `json:"id"`
    Title        string          `json:"title"`
    Description  *string         `json:"description,omitempty"`
    Location     *string         `json:"location,omitempty"`
    StartTime    time.Time       `json:"startTime"`
    EndTime      time.Time       `json:"endTime"`
    AllDay       bool            `json:"allDay"`
    CreatedBy    int             `json:"-"`
    AssigneeID   int             `json:"assigneeId"`
    Assignee     *models.Profile `json:"assignee,omitempty"`
    SourceModule string          `json:"sourceModule"`
    SourceID     *int            `json:"sourceId,omitempty"`
    FamilyID     int             `json:"familyId"`
    EventType    EventType       `json:"eventType"`
    
    IsRecurring       bool           `json:"isRecurring"`
    RecurrenceType    RecurrenceType `json:"recurrenceType"`
    RecurrenceEndTime *time.Time     `json:"recurrenceEndTime,omitempty"`
    IsException       bool           `json:"isException"` 
    ParentEventID     *int           `json:"parentEventId,omitempty"` 
    
    InstanceDate      *time.Time     `json:"instanceDate,omitempty"`
    
    CreatedAt    time.Time   `json:"createdAt"`
    UpdatedAt    time.Time   `json:"updatedAt"`
    IsDeleted    bool        `json:"isDeleted"`
    DeletedAt    *time.Time  `json:"deletedAt,omitempty"`
    DeletedBy    *int        `json:"deletedBy,omitempty"`
}