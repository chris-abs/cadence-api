package chores

import (
	"fmt"
	"time"
)

type CalendarService interface {
	CreateEvent(sourceModule string, sourceID int, title, description string, startTime, endTime time.Time, assigneeID, familyID int) error
	UpdateEvent(sourceModule string, sourceID int, title, description string, startTime, endTime time.Time, assigneeID, familyID int) error
	DeleteEvent(sourceModule string, sourceID int) error
}

type Service struct {
	repo            *Repository
	calendarService CalendarService
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) SetCalendarService(calendarService CalendarService) {
	s.calendarService = calendarService
}

func (s *Service) CreateChore(userID int, familyID int, req *CreateChoreRequest) (*Chore, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("chore name is required")
	}

	if req.AssigneeID == 0 {
		return nil, fmt.Errorf("assignee is required")
	}

	chore := &Chore{
		Name:           req.Name,
		Description:    req.Description,
		CreatorID:      userID,
		AssigneeID:     req.AssigneeID,
		FamilyID:       familyID,
		Points:         req.Points,
		OccurrenceType: req.OccurrenceType,
		OccurrenceData: req.OccurrenceData,
	}

	if err := s.repo.CreateChore(chore); err != nil {
		return nil, fmt.Errorf("failed to create chore: %v", err)
	}

	if !chore.OccurrenceData.StartDate.After(time.Now()) {
		if err := s.generateInitialInstances(chore); err != nil {
			fmt.Printf("Warning: failed to generate initial instances: %v\n", err)
		}
	}

	fullChore, err := s.repo.GetChoreByID(chore.ID, familyID)
	if err != nil {
		return nil, fmt.Errorf("chore created but failed to retrieve it: %v", err)
	}

	return fullChore, nil
}

func (s *Service) GetChoreByID(id int, familyID int) (*Chore, error) {
	return s.repo.GetChoreByID(id, familyID)
}

func (s *Service) GetChoresByFamilyID(familyID int) ([]*Chore, error) {
	return s.repo.GetChoresByFamilyID(familyID)
}

func (s *Service) GetChoresByAssigneeID(assigneeID int, familyID int) ([]*Chore, error) {
	return s.repo.GetChoresByAssigneeID(assigneeID, familyID)
}

func (s *Service) UpdateChore(id int, familyID int, req *UpdateChoreRequest) (*Chore, error) {
	chore, err := s.repo.GetChoreByID(id, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chore: %v", err)
	}

	chore.Name = req.Name
	chore.Description = req.Description
	chore.AssigneeID = req.AssigneeID
	chore.Points = req.Points
	chore.OccurrenceType = req.OccurrenceType
	chore.OccurrenceData = req.OccurrenceData

	if err := s.repo.UpdateChore(chore); err != nil {
		return nil, fmt.Errorf("failed to update chore: %v", err)
	}

	if s.calendarService != nil {
		// Todo: update when we have calendar structure - this should be used to update calendar event
	}

	updatedChore, err := s.repo.GetChoreByID(id, familyID)
	if err != nil {
		return nil, fmt.Errorf("chore updated but failed to retrieve it: %v", err)
	}

	return updatedChore, nil
}

func (s *Service) DeleteChore(id int, familyID int) error {
	chore, err := s.repo.GetChoreByID(id, familyID)
	if err != nil {
		return fmt.Errorf("failed to get chore: %v", err)
	}

	if s.calendarService != nil {
		for _, instance := range chore.Instances {
			if err := s.calendarService.DeleteEvent("chores", instance.ID); err != nil {
				fmt.Printf("Warning: failed to delete calendar event: %v\n", err)
			}
		}
	}

	if err := s.repo.DeleteChore(id, familyID); err != nil {
		return fmt.Errorf("failed to delete chore: %v", err)
	}

	return nil
}

func (s *Service) GetInstanceByID(id int, familyID int) (*ChoreInstance, error) {
	return s.repo.GetInstanceByID(id, familyID)
}

func (s *Service) GetInstancesByDueDate(date time.Time, familyID int) ([]*ChoreInstance, error) {
	return s.repo.GetInstancesByDueDate(date, familyID)
}

func (s *Service) GetInstancesByAssignee(assigneeID int, familyID int, startDate, endDate time.Time) ([]*ChoreInstance, error) {
	return s.repo.GetInstancesByAssignee(assigneeID, familyID, startDate, endDate)
}

func (s *Service) CompleteChoreInstance(id int, userID int, familyID int, req *UpdateChoreInstanceRequest) (*ChoreInstance, error) {
	instance, err := s.repo.GetInstanceByID(id, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chore instance: %v", err)
	}

	if instance.AssigneeID != userID {
		return nil, fmt.Errorf("only the assignee can mark this chore as completed")
	}

	now := time.Now().UTC()
	instance.Status = StatusCompleted
	instance.CompletedAt = &now
	instance.Notes = req.Notes
	
	if err := s.repo.UpdateChoreInstance(instance); err != nil {
		return nil, fmt.Errorf("failed to update chore instance: %v", err)
	}

	if s.calendarService != nil {
		// Todo: update when we have calendar structure - this should be used to update calendar event
	}

	return s.repo.GetInstanceByID(id, familyID)
}

func (s *Service) VerifyChoreInstance(id int, parentID int, familyID int, req *VerifyChoreInstanceRequest) (*ChoreInstance, error) {
	instance, err := s.repo.GetInstanceByID(id, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chore instance: %v", err)
	}

	if instance.Status != StatusCompleted {
		return nil, fmt.Errorf("only completed chores can be verified")
	}

	instance.Status = StatusVerified
	instance.VerifiedBy = &parentID
	if req.Notes != "" {
		instance.Notes = req.Notes
	}
	
	if err := s.repo.UpdateChoreInstance(instance); err != nil {
		return nil, fmt.Errorf("failed to verify chore instance: %v", err)
	}

	if s.calendarService != nil {
		// Todo: update when we have calendar structure - this should be used to update calendar event
	}

	return s.repo.GetInstanceByID(id, familyID)
}

func (s *Service) GetChoreStats(userID int, familyID int, startDate, endDate time.Time) (*ChoreStats, error) {
	return s.repo.GetChoreStats(userID, familyID, startDate, endDate)
}

func (s *Service) GenerateDailyChoreInstances(familyID int) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	
	chores, err := s.repo.GetChoresByFamilyID(familyID)
	if err != nil {
		return fmt.Errorf("failed to get chores: %v", err)
	}

	for _, chore := range chores {
		if s.shouldCreateInstanceForDate(chore, today) {
			exists, err := s.repo.CheckInstanceExists(chore.ID, today)
			if err != nil {
				fmt.Printf("Error checking if instance exists: %v\n", err)
				continue
			}

			if !exists {
				instance := &ChoreInstance{
					ChoreID:    chore.ID,
					AssigneeID: chore.AssigneeID,
					FamilyID:   chore.FamilyID,
					DueDate:    today,
					Status:     StatusPending,
				}

				if err := s.repo.CreateChoreInstance(instance); err != nil {
					fmt.Printf("Error creating chore instance: %v\n", err)
					continue
				}

				if s.calendarService != nil {
					err := s.calendarService.CreateEvent(
						"chores", 
						instance.ID,
						chore.Name,
						chore.Description,
						today,
						today.Add(1 * time.Hour), 
						chore.AssigneeID,
						chore.FamilyID,
					)
					if err != nil {
						fmt.Printf("Error creating calendar event: %v\n", err)
					}
				}
			}
		}
	}

	return nil
}

func (s *Service) generateInitialInstances(chore *Chore) error {
	startDate := chore.OccurrenceData.StartDate.Truncate(24 * time.Hour)
	today := time.Now().UTC().Truncate(24 * time.Hour)

	endDate := today
	if chore.OccurrenceData.EndDate != nil && chore.OccurrenceData.EndDate.Before(today) {
		endDate = *chore.OccurrenceData.EndDate
	}

	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		if s.shouldCreateInstanceForDate(chore, date) {
			exists, err := s.repo.CheckInstanceExists(chore.ID, date)
			if err != nil {
				return fmt.Errorf("error checking if instance exists: %v", err)
			}

			if !exists {
				instance := &ChoreInstance{
					ChoreID:    chore.ID,
					AssigneeID: chore.AssigneeID,
					FamilyID:   chore.FamilyID,
					DueDate:    date,
					Status:     StatusPending,
				}

				if err := s.repo.CreateChoreInstance(instance); err != nil {
					return fmt.Errorf("error creating chore instance: %v", err)
				}

				if s.calendarService != nil {
					err := s.calendarService.CreateEvent(
						"chores", 
						instance.ID,
						chore.Name,
						chore.Description,
						date,
						date.Add(1 * time.Hour),
						chore.AssigneeID,
						chore.FamilyID,
					)
					if err != nil {
						fmt.Printf("Error creating calendar event: %v\n", err)
					}
				}
			}
		}
	}

	return nil
}

func (s *Service) shouldCreateInstanceForDate(chore *Chore, date time.Time) bool {
	if date.Before(chore.OccurrenceData.StartDate) {
		return false
	}

	if chore.OccurrenceData.EndDate != nil && date.After(*chore.OccurrenceData.EndDate) {
		return false
	}

	switch chore.OccurrenceType {
	case OccurrenceDaily:
		return true

	case OccurrenceWeekly:
		weekday := date.Weekday()
		for _, day := range chore.OccurrenceData.DaysOfWeek {
			if weekday == day {
				return true
			}
		}
		return false

	case OccurrenceMonthly:
		dayOfMonth := date.Day()
		for _, day := range chore.OccurrenceData.DaysOfMonth {
			if dayOfMonth == day {
				return true
			}
		}
		return false

	case OccurrenceCustom:
		startDate := chore.OccurrenceData.StartDate.Truncate(24 * time.Hour)
		dateTruncated := date.Truncate(24 * time.Hour)
		
		daysDiff := int(dateTruncated.Sub(startDate).Hours() / 24)
		
		switch chore.OccurrenceData.IntervalUnit {
		case "day":
			return daysDiff%chore.OccurrenceData.Interval == 0
		case "week":
			weeksDiff := daysDiff / 7
			return weeksDiff%chore.OccurrenceData.Interval == 0
		case "month":
			monthsDiff := daysDiff / 30
			return monthsDiff%chore.OccurrenceData.Interval == 0
		}
	}

	return false
}