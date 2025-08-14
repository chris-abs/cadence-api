package entities

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type RecurrenceType string

const (
    RecurrenceDaily   RecurrenceType = "DAILY"
    RecurrenceWeekly  RecurrenceType = "WEEKLY"
    RecurrenceMonthly RecurrenceType = "MONTHLY"
    RecurrenceYearly  RecurrenceType = "YEARLY"
)

type Event struct {
    ID           models.EventID    `json:"id"`
    Title        string            `json:"title"`
    Description  *string           `json:"description,omitempty"`
    Location     *string           `json:"location,omitempty"`
    StartTime    time.Time         `json:"startTime"`
    EndTime      time.Time         `json:"endTime"`
    AllDay       bool              `json:"allDay"`
    CreatedBy    models.ProfileID  `json:"-"`
    AssigneeID   *models.ProfileID `json:"assigneeId,omitempty"`
    Assignee     *models.Profile   `json:"assignee,omitempty"`
    SourceModule string            `json:"sourceModule"`
    SourceID     *string           `json:"sourceId,omitempty"` 
    FamilyID     models.FamilyID   `json:"familyId"`
    
    IsRecurring       bool            `json:"isRecurring"`
    RecurrenceType    *RecurrenceType `json:"recurrenceType,omitempty"` 
    RecurrenceEndTime *time.Time      `json:"recurrenceEndTime,omitempty"`
    IsException       bool            `json:"isException"` 
    ParentEventID     *models.EventID `json:"parentEventId,omitempty"` 
    
    InstanceDate      *time.Time      `json:"instanceDate,omitempty"`
    
    CreatedAt    time.Time            `json:"createdAt"`
    UpdatedAt    time.Time            `json:"updatedAt"`
    IsDeleted    bool                 `json:"isDeleted"`
    DeletedAt    *time.Time           `json:"deletedAt,omitempty"`
    DeletedBy    *models.ProfileID    `json:"deletedBy,omitempty"`
}

func (rt *RecurrenceType) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    
    switch RecurrenceType(s) {
    case RecurrenceDaily, RecurrenceWeekly, RecurrenceMonthly, RecurrenceYearly:
        *rt = RecurrenceType(s)
        return nil
    default:
        return fmt.Errorf("invalid recurrence type: %s", s)
    }
}