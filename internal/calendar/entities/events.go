package entities

import "time"

type RecurrenceType string

const (
    RecurrenceNone    RecurrenceType = ""
    RecurrenceDaily   RecurrenceType = "DAILY"
    RecurrenceWeekly  RecurrenceType = "WEEKLY"
    RecurrenceMonthly RecurrenceType = "MONTHLY"
    RecurrenceYearly  RecurrenceType = "YEARLY"
)

type RecurrenceState string

const (
    RecurrenceActive    RecurrenceState = "ACTIVE"
    RecurrenceCancelled RecurrenceState = "CANCELLED"
)

type Event struct {
    ID           int        `json:"id"`
    Title        string     `json:"title"`
    Description  *string    `json:"description,omitempty"`
    Location     *string    `json:"location,omitempty"`
    StartTime    time.Time  `json:"startTime"`
    EndTime      time.Time  `json:"endTime"`
    AllDay       bool       `json:"allDay"`
    CreatedBy    int        `json:"createdBy"`
    AssigneeID   int        `json:"assigneeId"`
    Type         string     `json:"type"`
    SourceModule string     `json:"sourceModule"`
    SourceID     *int       `json:"sourceId,omitempty"`
    FamilyID     int        `json:"familyId"`
    
    RecurrenceType      RecurrenceType  `json:"recurrenceType"`
    RecurrenceInterval  *int            `json:"recurrenceInterval,omitempty"`
    RecurrenceEndTime   *time.Time      `json:"recurrenceEndTime,omitempty"`
    RecurrenceState     RecurrenceState `json:"recurrenceState,omitempty"`
    RecurrenceCancelledAt *time.Time    `json:"recurrenceCancelledAt,omitempty"`
    
    CreatedAt    time.Time  `json:"createdAt"`
    UpdatedAt    time.Time  `json:"updatedAt"`
    IsDeleted    bool       `json:"isDeleted"`
    DeletedAt    *time.Time `json:"deletedAt,omitempty"`
    DeletedBy    *int       `json:"deletedBy,omitempty"`
}

func (e *Event) IsRecurring() bool {
    return e.RecurrenceType != RecurrenceNone
}

func (e *Event) IsRecurrenceActive() bool {
    return e.IsRecurring() && e.RecurrenceState == RecurrenceActive
}