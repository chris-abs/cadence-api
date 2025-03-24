package entities

import (
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
	DaysOfWeek []time.Weekday `json:"daysOfWeek,omitempty"`
	DaysOfMonth []int `json:"daysOfMonth,omitempty"`
	StartDate   time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate,omitempty"`
	Interval    int `json:"interval,omitempty"`
	IntervalUnit string `json:"intervalUnit,omitempty"`
}

type Chore struct {
	ID             int           `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	CreatorID      int           `json:"creatorId"`
	AssigneeID     int           `json:"assigneeId"`
	FamilyID       int           `json:"familyId"`
	Points         int           `json:"points"`
	OccurrenceType OccurrenceType `json:"occurrenceType"`
	OccurrenceData OccurrenceData `json:"occurrenceData"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
	
	Assignee      *models.Profile    `json:"assignee,omitempty"`
	Creator       *models.Profile    `json:"creator,omitempty"`
	Instances     []ChoreInstance `json:"instances,omitempty"`
}

type ChoreInstance struct {
	ID           int         `json:"id"`
	ChoreID      int         `json:"choreId"`
	AssigneeID   int         `json:"assigneeId"`
	FamilyID     int         `json:"familyId"`
	DueDate      time.Time   `json:"dueDate"`
	Status       ChoreStatus `json:"status"`
	CompletedAt  *time.Time  `json:"completedAt,omitempty"`
	VerifiedBy   *int        `json:"verifiedBy,omitempty"`
	Notes        string      `json:"notes"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	
	Chore        *Chore       `json:"chore,omitempty"`
	Assignee     *models.Profile `json:"assignee,omitempty"`
	Verifier     *models.Profile `json:"verifier,omitempty"`
}

type DailyVerification struct {
	Date          time.Time   `json:"date"`
	AssigneeID    int         `json:"assigneeId"`
	FamilyID      int         `json:"familyId"`
	IsVerified    bool        `json:"isVerified"`
	VerifiedBy    *int        `json:"verifiedBy,omitempty"`
	VerifiedAt    *time.Time  `json:"verifiedAt,omitempty"`
	Notes         string      `json:"notes"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}