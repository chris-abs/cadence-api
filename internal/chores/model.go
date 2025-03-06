package chores

import (
	"time"

	"github.com/chrisabs/cadence/internal/chores/entities"
)

type CreateChoreRequest struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	AssigneeID     int                    `json:"assigneeId"`
	Points         int                    `json:"points"`
	OccurrenceType entities.OccurrenceType `json:"occurrenceType"`
	OccurrenceData entities.OccurrenceData `json:"occurrenceData"`
}

type UpdateChoreRequest struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	AssigneeID     int                    `json:"assigneeId"`
	Points         int                    `json:"points"`
	OccurrenceType entities.OccurrenceType `json:"occurrenceType"`
	OccurrenceData entities.OccurrenceData `json:"occurrenceData"`
}

type UpdateChoreInstanceRequest struct {
	Status      entities.ChoreStatus `json:"status"`
	Notes       string              `json:"notes"`
}

type ReviewChoreRequest struct {
	Status      entities.ChoreStatus `json:"status"`
	Notes       string              `json:"notes"`
}

type VerifyDayRequest struct {
	Date        string `json:"date"`
	AssigneeID  int    `json:"assigneeId"`
	Notes       string `json:"notes"`
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