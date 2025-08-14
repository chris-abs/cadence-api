package entities

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type OccurrenceType string

const (
    OccurrenceDaily   OccurrenceType = "daily"
    OccurrenceWeekly  OccurrenceType = "weekly"
    OccurrenceMonthly OccurrenceType = "monthly"
    OccurrenceCustom  OccurrenceType = "custom"
)

type ChoreStatus string

const (
    StatusPending   ChoreStatus = "pending"
    StatusCompleted ChoreStatus = "completed"
    StatusVerified  ChoreStatus = "verified"
    StatusRejected  ChoreStatus = "rejected"
    StatusMissed    ChoreStatus = "missed"
)

type OccurrenceData struct {
    DaysOfWeek   []time.Weekday `json:"daysOfWeek,omitempty"`
    DaysOfMonth  []int          `json:"daysOfMonth,omitempty"`
    StartDate    time.Time      `json:"startDate"`
    EndDate      *time.Time     `json:"endDate,omitempty"`
    Interval     int            `json:"interval,omitempty"`
    IntervalUnit string         `json:"intervalUnit,omitempty"`
}

func (od *OccurrenceData) UnmarshalJSON(data []byte) error {
    type Alias OccurrenceData
    aux := &struct {
        StartDate string  `json:"startDate"`
        EndDate   *string `json:"endDate,omitempty"`
        *Alias
    }{
        Alias: (*Alias)(od),
    }
    
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    
    startDate, err := time.Parse(time.RFC3339, aux.StartDate)
    if err != nil {
        startDate, err = time.Parse("2006-01-02", aux.StartDate)
        if err != nil {
            return fmt.Errorf("invalid startDate format: %v", err)
        }
    }
    od.StartDate = startDate.UTC()
    
    if aux.EndDate != nil {
        endDate, err := time.Parse(time.RFC3339, *aux.EndDate)
        if err != nil {
            endDate, err = time.Parse("2006-01-02", *aux.EndDate)
            if err != nil {
                return fmt.Errorf("invalid endDate format: %v", err)
            }
        }
        endDateUTC := endDate.UTC()
        od.EndDate = &endDateUTC
    }
    
    return nil
}

func (od *OccurrenceData) MarshalJSON() ([]byte, error) {
    type Alias OccurrenceData
    return json.Marshal(&struct {
        StartDate string  `json:"startDate"`
        EndDate   *string `json:"endDate,omitempty"`
        *Alias
    }{
        StartDate: od.StartDate.Format(time.RFC3339), 
        EndDate: func() *string {
            if od.EndDate != nil {
                s := od.EndDate.Format(time.RFC3339)
                return &s
            }
            return nil
        }(),
        Alias: (*Alias)(od),
    })
}

type Chore struct {
    ID             models.ChoreID       `json:"id"`
    Name           string               `json:"name"`
    Description    string               `json:"description"`
    CreatorID      models.ProfileID     `json:"creatorId"`
    AssigneeID     models.ProfileID     `json:"assigneeId"`
    FamilyID       models.FamilyID      `json:"familyId"`
    Points         int                  `json:"points"`
    OccurrenceType OccurrenceType       `json:"occurrenceType"`
    OccurrenceData OccurrenceData       `json:"occurrenceData"`
    CreatedAt      time.Time            `json:"createdAt"`
    UpdatedAt      time.Time            `json:"updatedAt"`
    
    Assignee      *models.Profile       `json:"assignee,omitempty"`
    Creator       *models.Profile       `json:"creator,omitempty"`
    Instances     []ChoreInstance       `json:"instances,omitempty"`
}

type ChoreInstance struct {
    ID              models.ChoreInstanceID `json:"id"`
    ChoreID         models.ChoreID         `json:"choreId"`
    AssigneeID      models.ProfileID       `json:"assigneeId"`
    FamilyID        models.FamilyID        `json:"familyId"`
    DueDate         time.Time              `json:"dueDate"`
    Status          ChoreStatus            `json:"status"`
    CompletedAt     *time.Time             `json:"completedAt,omitempty"`
    VerifiedBy      *models.ProfileID      `json:"verifiedBy,omitempty"`
    VerifiedAt      *time.Time             `json:"verifiedAt,omitempty"`
    RejectionReason string                 `json:"rejectionReason,omitempty"`
    Notes           string                 `json:"notes"`
    CreatedAt       time.Time              `json:"createdAt"`
    UpdatedAt       time.Time              `json:"updatedAt"`
    IsDeleted       bool                   `json:"isDeleted"`
    DeletedAt       *time.Time             `json:"deletedAt,omitempty"`
    DeletedBy       *models.ProfileID      `json:"deletedBy,omitempty"`
    
    Chore          *Chore                  `json:"chore,omitempty"`
    Assignee       *models.Profile         `json:"assignee,omitempty"`
    Verifier       *models.Profile         `json:"verifier,omitempty"`
}

type DailyVerification struct {
    Date          time.Time            `json:"date"`
    AssigneeID    models.ProfileID     `json:"assigneeId"`
    FamilyID      models.FamilyID      `json:"familyId"`
    IsVerified    bool                 `json:"isVerified"`
    VerifiedBy    *models.ProfileID    `json:"verifiedBy,omitempty"`
    VerifiedAt    *time.Time           `json:"verifiedAt,omitempty"`
    Notes         string               `json:"notes"`
    CreatedAt     time.Time            `json:"createdAt"`
    UpdatedAt     time.Time            `json:"updatedAt"`
}