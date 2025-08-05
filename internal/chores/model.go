package chores

import (
	"time"

	"github.com/chrisabs/cadence/internal/chores/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type CreateChoreRequest struct {
	Name           string                  `json:"name"`
	Description    string                  `json:"description"`
	AssigneeID     models.ProfileID        `json:"assigneeId"`
	Points         int                     `json:"points"`
	OccurrenceType entities.OccurrenceType `json:"occurrenceType"`
	OccurrenceData entities.OccurrenceData `json:"occurrenceData"`
}

type UpdateChoreRequest struct {
	Name           string                  `json:"name"`
	Description    string                  `json:"description"`
	AssigneeID     models.ProfileID        `json:"assigneeId"`
	Points         int                     `json:"points"`
	OccurrenceType entities.OccurrenceType `json:"occurrenceType"`
	OccurrenceData entities.OccurrenceData `json:"occurrenceData"`
}

type UpdateChoreInstanceRequest struct {
	Status      entities.ChoreStatus `json:"status"`
	Notes       string               `json:"notes"`
}

type ReviewChoreRequest struct {
    Status          entities.ChoreStatus `json:"status"`
    Notes           string               `json:"notes"`
    RejectionReason string               `json:"rejectionReason,omitempty"` 
}

type VerifyDayRequest struct {
    Date            string               `json:"date"`
    AssigneeID      models.ProfileID     `json:"assigneeId"`
    Status          entities.ChoreStatus `json:"status"`
    Notes           string               `json:"notes"`
    RejectionReason string               `json:"rejectionReason,omitempty"`
}

type ChoreVerificationResponse struct {
    Success        bool                    `json:"success"`
    Message        string                  `json:"message"`
    VerifiedCount  int                     `json:"verifiedCount,omitempty"`
    RejectedCount  int                     `json:"rejectedCount,omitempty"`
    Instance       *entities.ChoreInstance `json:"instance,omitempty"`
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
	ProfileId    models.ProfileID `json:"profileId"`
	StartDate    time.Time        `json:"startDate"`
	EndDate      time.Time        `json:"endDate"`
}