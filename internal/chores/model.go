package chores

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
	DaysOfWeek []time.Weekday  `json:"daysOfWeek,omitempty"`
	
	DaysOfMonth []int          `json:"daysOfMonth,omitempty"`
	
	StartDate   time.Time      `json:"startDate"`
	EndDate     *time.Time     `json:"endDate,omitempty"`
	
	Interval    int            `json:"interval,omitempty"`
	IntervalUnit string        `json:"intervalUnit,omitempty"` 
}

type Chore struct {
	ID             int                 `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	CreatorID      int                 `json:"creatorId"`
	AssigneeID     int                 `json:"assigneeId"`
	FamilyID       int                 `json:"familyId"`
	Points         int                 `json:"points"`
	OccurrenceType OccurrenceType      `json:"occurrenceType"`
	OccurrenceData OccurrenceData      `json:"occurrenceData"`
	CreatedAt      time.Time           `json:"createdAt"`
	UpdatedAt      time.Time           `json:"updatedAt"`
	
	Assignee      *models.User         `json:"assignee,omitempty"`
	Creator       *models.User         `json:"creator,omitempty"`
	Instances     []ChoreInstance      `json:"instances,omitempty"`
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
	Assignee     *models.User `json:"assignee,omitempty"`
	Verifier     *models.User `json:"verifier,omitempty"`
}

type CreateChoreRequest struct {
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	AssigneeID     int           `json:"assigneeId"`
	Points         int           `json:"points"`
	OccurrenceType OccurrenceType `json:"occurrenceType"`
	OccurrenceData OccurrenceData `json:"occurrenceData"`
}

type UpdateChoreRequest struct {
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	AssigneeID     int           `json:"assigneeId"`
	Points         int           `json:"points"`
	OccurrenceType OccurrenceType `json:"occurrenceType"`
	OccurrenceData OccurrenceData `json:"occurrenceData"`
}

type UpdateChoreInstanceRequest struct {
	Status      ChoreStatus `json:"status"`
	Notes       string      `json:"notes"`
}

type VerifyChoreInstanceRequest struct {
	Notes       string     `json:"notes"`
}

type ChoreStats struct {
	TotalAssigned     int     `json:"totalAssigned"`
	TotalCompleted    int     `json:"totalCompleted"`
	TotalVerified     int     `json:"totalVerified"`
	TotalMissed       int     `json:"totalMissed"`
	CompletionRate    float64 `json:"completionRate"`
	PointsEarned      int     `json:"pointsEarned"`
}

type ChoreStatsRequest struct {
	UserID    int        `json:"userId"`
	StartDate time.Time  `json:"startDate"`
	EndDate   time.Time  `json:"endDate"`
}