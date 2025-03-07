package chores

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/chores/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateChore(chore *entities.Chore) error {
	occurrenceData, err := json.Marshal(chore.OccurrenceData)
	if err != nil {
		return fmt.Errorf("error marshaling occurrence data: %v", err)
	}

	query := `
		INSERT INTO chore (
			name, description, creator_id, assignee_id, family_id,
			points, occurrence_type, occurrence_data, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	err = r.db.QueryRow(
		query,
		chore.Name,
		chore.Description,
		chore.CreatorID,
		chore.AssigneeID,
		chore.FamilyID,
		chore.Points,
		chore.OccurrenceType,
		occurrenceData,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&chore.ID)

	if err != nil {
		return fmt.Errorf("error creating chore: %v", err)
	}

	return nil
}

func (r *Repository) GetChoreByID(id int, familyID int) (*entities.Chore, error) {
    query := `
        SELECT c.id, c.name, c.description, c.creator_id, c.assignee_id, c.family_id,
               c.points, c.occurrence_type, c.occurrence_data, c.created_at, c.updated_at,
               creator.id, creator.email, creator.first_name, creator.last_name, creator.image_url,
               assignee.id, assignee.email, assignee.first_name, assignee.last_name, assignee.image_url
        FROM chore c
        LEFT JOIN users creator ON c.creator_id = creator.id AND creator.is_deleted = false
        LEFT JOIN users assignee ON c.assignee_id = assignee.id AND assignee.is_deleted = false
        WHERE c.id = $1 AND c.family_id = $2 AND c.is_deleted = false`

	chore := &entities.Chore{}
	creator := &models.User{}
	assignee := &models.User{}
	var occurrenceDataJSON []byte

	err := r.db.QueryRow(query, id, familyID).Scan(
		&chore.ID, &chore.Name, &chore.Description, &chore.CreatorID, &chore.AssigneeID, &chore.FamilyID,
		&chore.Points, &chore.OccurrenceType, &occurrenceDataJSON, &chore.CreatedAt, &chore.UpdatedAt,
		&creator.ID, &creator.Email, &creator.FirstName, &creator.LastName, &creator.ImageURL,
		&assignee.ID, &assignee.Email, &assignee.FirstName, &assignee.LastName, &assignee.ImageURL,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chore not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting chore: %v", err)
	}

	if err := json.Unmarshal(occurrenceDataJSON, &chore.OccurrenceData); err != nil {
		return nil, fmt.Errorf("error unmarshaling occurrence data: %v", err)
	}

	chore.Creator = creator
	chore.Assignee = assignee

	instances, err := r.GetInstancesByChoreID(chore.ID, familyID)
	if err != nil {
		return nil, fmt.Errorf("error getting chore instances: %v", err)
	}
	chore.Instances = instances

	return chore, nil
}

func (r *Repository) GetChoresByFamilyID(familyID int) ([]*entities.Chore, error) {
    query := `
        SELECT c.id, c.name, c.description, c.creator_id, c.assignee_id, c.family_id,
               c.points, c.occurrence_type, c.occurrence_data, c.created_at, c.updated_at,
               creator.id, creator.email, creator.first_name, creator.last_name, creator.image_url,
               assignee.id, assignee.email, assignee.first_name, assignee.last_name, assignee.image_url
        FROM chore c
        LEFT JOIN users creator ON c.creator_id = creator.id AND creator.is_deleted = false
        LEFT JOIN users assignee ON c.assignee_id = assignee.id AND assignee.is_deleted = false
        WHERE c.family_id = $1 AND c.is_deleted = false
        ORDER BY c.created_at DESC`

	rows, err := r.db.Query(query, familyID)
	if err != nil {
		return nil, fmt.Errorf("error getting chores: %v", err)
	}
	defer rows.Close()

	var chores []*entities.Chore
	for rows.Next() {
		chore := &entities.Chore{}
		creator := &models.User{}
		assignee := &models.User{}
		var occurrenceDataJSON []byte

		err := rows.Scan(
			&chore.ID, &chore.Name, &chore.Description, &chore.CreatorID, &chore.AssigneeID, &chore.FamilyID,
			&chore.Points, &chore.OccurrenceType, &occurrenceDataJSON, &chore.CreatedAt, &chore.UpdatedAt,
			&creator.ID, &creator.Email, &creator.FirstName, &creator.LastName, &creator.ImageURL,
			&assignee.ID, &assignee.Email, &assignee.FirstName, &assignee.LastName, &assignee.ImageURL,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning chore: %v", err)
		}

		if err := json.Unmarshal(occurrenceDataJSON, &chore.OccurrenceData); err != nil {
			return nil, fmt.Errorf("error unmarshaling occurrence data: %v", err)
		}

		chore.Creator = creator
		chore.Assignee = assignee
		chores = append(chores, chore)
	}

	return chores, nil
}

func (r *Repository) GetChoresByAssigneeID(assigneeID int, familyID int) ([]*entities.Chore, error) {
    query := `
        SELECT c.id, c.name, c.description, c.creator_id, c.assignee_id, c.family_id,
               c.points, c.occurrence_type, c.occurrence_data, c.created_at, c.updated_at
        FROM chore c
        WHERE c.assignee_id = $1 AND c.family_id = $2 AND c.is_deleted = false
        ORDER BY c.created_at DESC`

	rows, err := r.db.Query(query, assigneeID, familyID)
	if err != nil {
		return nil, fmt.Errorf("error getting chores: %v", err)
	}
	defer rows.Close()

	var chores []*entities.Chore
	for rows.Next() {
		chore := &entities.Chore{}
		var occurrenceDataJSON []byte

		err := rows.Scan(
			&chore.ID, &chore.Name, &chore.Description, &chore.CreatorID, &chore.AssigneeID, &chore.FamilyID,
			&chore.Points, &chore.OccurrenceType, &occurrenceDataJSON, &chore.CreatedAt, &chore.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning chore: %v", err)
		}

		if err := json.Unmarshal(occurrenceDataJSON, &chore.OccurrenceData); err != nil {
			return nil, fmt.Errorf("error unmarshaling occurrence data: %v", err)
		}

		chores = append(chores, chore)
	}

	return chores, nil
}

func (r *Repository) UpdateChore(chore *entities.Chore) error {
	occurrenceData, err := json.Marshal(chore.OccurrenceData)
	if err != nil {
		return fmt.Errorf("error marshaling occurrence data: %v", err)
	}

	query := `
		UPDATE chore
		SET name = $2, description = $3, assignee_id = $4, points = $5, 
			occurrence_type = $6, occurrence_data = $7, updated_at = $8
		WHERE id = $1 AND family_id = $9 AND is_deleted = false`

	result, err := r.db.Exec(
		query,
		chore.ID,
		chore.Name,
		chore.Description,
		chore.AssigneeID,
		chore.Points,
		chore.OccurrenceType,
		occurrenceData,
		time.Now().UTC(),
		chore.FamilyID,
	)

	if err != nil {
		return fmt.Errorf("error updating chore: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chore not found")
	}

	return nil
}

func (r *Repository) DeleteChore(id int, familyID int, deletedBy int) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    instanceQuery := `
        UPDATE chore_instance
        SET is_deleted = true, deleted_at = $3, deleted_by = $4, updated_at = $3
        WHERE chore_id = $1 AND family_id = $2 AND is_deleted = false`
    
    _, err = tx.Exec(instanceQuery, id, familyID, time.Now().UTC(), deletedBy)
    if err != nil {
        return fmt.Errorf("error soft deleting chore instances: %v", err)
    }

    choreQuery := `
        UPDATE chore
        SET is_deleted = true, deleted_at = $3, deleted_by = $4, updated_at = $3
        WHERE id = $1 AND family_id = $2 AND is_deleted = false`
    
    result, err := tx.Exec(choreQuery, id, familyID, time.Now().UTC(), deletedBy)
    if err != nil {
        return fmt.Errorf("error soft deleting chore: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking delete result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("chore not found")
    }

    return tx.Commit()
}

func (r *Repository) RestoreChore(id int, familyID int) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    choreQuery := `
        UPDATE chore
        SET is_deleted = false, deleted_at = NULL, deleted_by = NULL, updated_at = $3
        WHERE id = $1 AND family_id = $2 AND is_deleted = true`
    
    result, err := tx.Exec(choreQuery, id, familyID, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error restoring chore: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking restore result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("chore not found or not deleted")
    }

    instanceQuery := `
        UPDATE chore_instance
        SET is_deleted = false, deleted_at = NULL, deleted_by = NULL, updated_at = $3
        WHERE chore_id = $1 AND family_id = $2 AND is_deleted = true`
    
    _, err = tx.Exec(instanceQuery, id, familyID, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error restoring chore instances: %v", err)
    }

    return tx.Commit()
}

func (r *Repository) CreateChoreInstance(instance *entities.ChoreInstance) error {
	query := `
		INSERT INTO chore_instance (
			chore_id, assignee_id, family_id, due_date, status, 
			notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		instance.ChoreID,
		instance.AssigneeID,
		instance.FamilyID,
		instance.DueDate,
		instance.Status,
		instance.Notes,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&instance.ID, &instance.CreatedAt, &instance.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating chore instance: %v", err)
	}

	return nil
}

func (r *Repository) GetInstanceByID(id int, familyID int) (*entities.ChoreInstance, error) {
    query := `
        SELECT ci.id, ci.chore_id, ci.assignee_id, ci.family_id, ci.due_date,
               ci.status, ci.completed_at, ci.verified_by, ci.notes, 
               ci.created_at, ci.updated_at,
               a.id, a.email, a.first_name, a.last_name, a.image_url,
               v.id, v.email, v.first_name, v.last_name, v.image_url
        FROM chore_instance ci
        LEFT JOIN users a ON ci.assignee_id = a.id AND a.is_deleted = false
        LEFT JOIN users v ON ci.verified_by = v.id AND v.is_deleted = false
        WHERE ci.id = $1 AND ci.family_id = $2 AND ci.is_deleted = false`

	instance := &entities.ChoreInstance{}
	assignee := &models.User{}
	verifier := &models.User{}
	var verifiedBy sql.NullInt64
	var completedAt sql.NullTime

	err := r.db.QueryRow(query, id, familyID).Scan(
		&instance.ID, &instance.ChoreID, &instance.AssigneeID, &instance.FamilyID, &instance.DueDate,
		&instance.Status, &completedAt, &verifiedBy, &instance.Notes,
		&instance.CreatedAt, &instance.UpdatedAt,
		&assignee.ID, &assignee.Email, &assignee.FirstName, &assignee.LastName, &assignee.ImageURL,
		&verifier.ID, &verifier.Email, &verifier.FirstName, &verifier.LastName, &verifier.ImageURL,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chore instance not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting chore instance: %v", err)
	}

	instance.Assignee = assignee

	if completedAt.Valid {
		instance.CompletedAt = &completedAt.Time
	}

	if verifiedBy.Valid {
		vID := int(verifiedBy.Int64)
		instance.VerifiedBy = &vID
		instance.Verifier = verifier
	}

	chore, err := r.GetChoreByID(instance.ChoreID, familyID)
	if err != nil {
		return nil, fmt.Errorf("error getting chore: %v", err)
	}
	instance.Chore = chore

	return instance, nil
}

func (r *Repository) GetInstancesByChoreID(choreID int, familyID int) ([]entities.ChoreInstance, error) {
    query := `
        SELECT ci.id, ci.chore_id, ci.assignee_id, ci.family_id, ci.due_date,
               ci.status, ci.completed_at, ci.verified_by, ci.notes, 
               ci.created_at, ci.updated_at
        FROM chore_instance ci
        WHERE ci.chore_id = $1 AND ci.family_id = $2 AND ci.is_deleted = false
        ORDER BY ci.due_date DESC`

	rows, err := r.db.Query(query, choreID, familyID)
	if err != nil {
		return nil, fmt.Errorf("error getting chore instances: %v", err)
	}
	defer rows.Close()

	var instances []entities.ChoreInstance
	for rows.Next() {
		instance := entities.ChoreInstance{}
		var verifiedBy sql.NullInt64
		var completedAt sql.NullTime

		err := rows.Scan(
			&instance.ID, &instance.ChoreID, &instance.AssigneeID, &instance.FamilyID, &instance.DueDate,
			&instance.Status, &completedAt, &verifiedBy, &instance.Notes,
			&instance.CreatedAt, &instance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning chore instance: %v", err)
		}

		if completedAt.Valid {
			instance.CompletedAt = &completedAt.Time
		}

		if verifiedBy.Valid {
			vID := int(verifiedBy.Int64)
			instance.VerifiedBy = &vID
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

func (r *Repository) GetInstancesByDueDate(dueDate time.Time, familyID int) ([]*entities.ChoreInstance, error) {
    startOfDay := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
    endOfDay := startOfDay.Add(24 * time.Hour)

    query := `
        SELECT ci.id, ci.chore_id, ci.assignee_id, ci.family_id, ci.due_date,
               ci.status, ci.completed_at, ci.verified_by, ci.notes, 
               ci.created_at, ci.updated_at,
               c.name, c.points
        FROM chore_instance ci
        JOIN chore c ON ci.chore_id = c.id AND c.is_deleted = false
        WHERE ci.due_date >= $1 AND ci.due_date < $2 AND ci.family_id = $3 AND ci.is_deleted = false
        ORDER BY ci.due_date ASC`

	rows, err := r.db.Query(query, startOfDay, endOfDay, familyID)
	if err != nil {
		return nil, fmt.Errorf("error getting chore instances: %v", err)
	}
	defer rows.Close()

	var instances []*entities.ChoreInstance
	for rows.Next() {
		instance := &entities.ChoreInstance{}
		chore := &entities.Chore{}
		var verifiedBy sql.NullInt64
		var completedAt sql.NullTime

		err := rows.Scan(
			&instance.ID, &instance.ChoreID, &instance.AssigneeID, &instance.FamilyID, &instance.DueDate,
			&instance.Status, &completedAt, &verifiedBy, &instance.Notes,
			&instance.CreatedAt, &instance.UpdatedAt,
			&chore.Name, &chore.Points,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning chore instance: %v", err)
		}

		if completedAt.Valid {
			instance.CompletedAt = &completedAt.Time
		}

		if verifiedBy.Valid {
			vID := int(verifiedBy.Int64)
			instance.VerifiedBy = &vID
		}

		chore.ID = instance.ChoreID
		instance.Chore = chore
		instances = append(instances, instance)
	}

	return instances, nil
}

func (r *Repository) GetInstancesByAssignee(assigneeID int, familyID int, startDate, endDate time.Time) ([]*entities.ChoreInstance, error) {
	query := `
    SELECT ci.id, ci.chore_id, ci.assignee_id, ci.family_id, ci.due_date,
        ci.status, ci.completed_at, ci.verified_by, ci.notes, 
        ci.created_at, ci.updated_at,
        c.name, c.points
    FROM chore_instance ci
    JOIN chore c ON ci.chore_id = c.id AND c.is_deleted = false
    WHERE ci.assignee_id = $1 AND ci.family_id = $2 
    AND ci.due_date >= $3 AND ci.due_date < $4
    AND ci.is_deleted = false
    ORDER BY ci.due_date ASC`

	rows, err := r.db.Query(query, assigneeID, familyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error getting chore instances: %v", err)
	}
	defer rows.Close()

	var instances []*entities.ChoreInstance
	for rows.Next() {
		instance := &entities.ChoreInstance{}
		chore := &entities.Chore{}
		var verifiedBy sql.NullInt64
		var completedAt sql.NullTime

		err := rows.Scan(
			&instance.ID, &instance.ChoreID, &instance.AssigneeID, &instance.FamilyID, &instance.DueDate,
			&instance.Status, &completedAt, &verifiedBy, &instance.Notes,
			&instance.CreatedAt, &instance.UpdatedAt,
			&chore.Name, &chore.Points,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning chore instance: %v", err)
		}

		if completedAt.Valid {
			instance.CompletedAt = &completedAt.Time
		}

		if verifiedBy.Valid {
			vID := int(verifiedBy.Int64)
			instance.VerifiedBy = &vID
		}

		chore.ID = instance.ChoreID
		instance.Chore = chore
		instances = append(instances, instance)
	}

	return instances, nil
}

func (r *Repository) GetInstancesByAssigneeAndDate(assigneeID int, familyID int, date time.Time) ([]*entities.ChoreInstance, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT ci.id, ci.chore_id, ci.assignee_id, ci.family_id, ci.due_date,
			   ci.status, ci.completed_at, ci.verified_by, ci.notes, 
			   ci.created_at, ci.updated_at,
			   c.name, c.points
		FROM chore_instance ci
		JOIN chore c ON ci.chore_id = c.id
		WHERE ci.assignee_id = $1 AND ci.family_id = $2 
		AND ci.due_date >= $3 AND ci.due_date < $4
		ORDER BY ci.due_date ASC`

	rows, err := r.db.Query(query, assigneeID, familyID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("error getting chore instances: %v", err)
	}
	defer rows.Close()

	var instances []*entities.ChoreInstance
	for rows.Next() {
		instance := &entities.ChoreInstance{}
		chore := &entities.Chore{}
		var verifiedBy sql.NullInt64
		var completedAt sql.NullTime

		err := rows.Scan(
			&instance.ID, &instance.ChoreID, &instance.AssigneeID, &instance.FamilyID, &instance.DueDate,
			&instance.Status, &completedAt, &verifiedBy, &instance.Notes,
			&instance.CreatedAt, &instance.UpdatedAt,
			&chore.Name, &chore.Points,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning chore instance: %v", err)
		}

		if completedAt.Valid {
			instance.CompletedAt = &completedAt.Time
		}

		if verifiedBy.Valid {
			vID := int(verifiedBy.Int64)
			instance.VerifiedBy = &vID
		}

		chore.ID = instance.ChoreID
		instance.Chore = chore
		instances = append(instances, instance)
	}

	return instances, nil
}

func (r *Repository) CheckInstanceExists(choreID int, dueDate time.Time) (bool, error) {
	startOfDay := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM chore_instance 
			WHERE chore_id = $1 AND due_date >= $2 AND due_date < $3
		)`

	var exists bool
	err := r.db.QueryRow(query, choreID, startOfDay, endOfDay).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if instance exists: %v", err)
	}

	return exists, nil
}

func (r *Repository) UpdateChoreInstance(instance *entities.ChoreInstance) error {
    query := `
        UPDATE chore_instance
        SET status = $2, completed_at = $3, verified_by = $4, notes = $5, updated_at = $6
        WHERE id = $1 AND family_id = $7 AND is_deleted = false`

	var completedAt *time.Time
	var verifiedBy *int

	if instance.Status == entities.StatusCompleted || instance.Status == entities.StatusVerified {
		if instance.CompletedAt == nil {
			now := time.Now().UTC()
			completedAt = &now
		} else {
			completedAt = instance.CompletedAt
		}
	}

	if instance.Status == entities.StatusVerified {
		verifiedBy = instance.VerifiedBy
	}

	result, err := r.db.Exec(
		query,
		instance.ID,
		instance.Status,
		completedAt,
		verifiedBy,
		instance.Notes,
		time.Now().UTC(),
		instance.FamilyID,
	)

	if err != nil {
		return fmt.Errorf("error updating chore instance: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chore instance not found")
	}

	return nil
}

func (r *Repository) GetChoreStats(userID int, familyID int, startDate, endDate time.Time) (*ChoreStats, error) {
	query := `
    SELECT 
        COUNT(*) as total_assigned,
        SUM(CASE WHEN status = 'completed' OR status = 'verified' THEN 1 ELSE 0 END) as total_completed,
        SUM(CASE WHEN status = 'verified' THEN 1 ELSE 0 END) as total_verified,
        SUM(CASE WHEN status = 'missed' THEN 1 ELSE 0 END) as total_missed,
        SUM(CASE WHEN status = 'completed' OR status = 'verified' THEN c.points ELSE 0 END) as points_earned
    FROM chore_instance ci
    JOIN chore c ON ci.chore_id = c.id AND c.is_deleted = false
    WHERE ci.assignee_id = $1 AND ci.family_id = $2 
    AND ci.due_date >= $3 AND ci.due_date <= $4
    AND ci.is_deleted = false`


	stats := &ChoreStats{}
	var totalCompleted, totalVerified, totalMissed sql.NullInt64
	var pointsEarned sql.NullInt64

	err := r.db.QueryRow(query, userID, familyID, startDate, endDate).Scan(
		&stats.TotalAssigned,
		&totalCompleted,
		&totalVerified,
		&totalMissed,
		&pointsEarned,
	)

	if err != nil {
		return nil, fmt.Errorf("error getting chore stats: %v", err)
	}

	if totalCompleted.Valid {
		stats.TotalCompleted = int(totalCompleted.Int64)
	}
	if totalVerified.Valid {
		stats.TotalVerified = int(totalVerified.Int64)
	}
	if totalMissed.Valid {
		stats.TotalMissed = int(totalMissed.Int64)
	}
	if pointsEarned.Valid {
		stats.PointsEarned = int(pointsEarned.Int64)
	}

	if stats.TotalAssigned > 0 {
		stats.CompletionRate = float64(stats.TotalCompleted) / float64(stats.TotalAssigned) * 100
	}

	return stats, nil
}

func (r *Repository) SaveDailyVerification(verification *entities.DailyVerification) error {
	existingQuery := `
		SELECT EXISTS(
			SELECT 1 FROM daily_verification 
			WHERE date = $1 AND assignee_id = $2 AND family_id = $3
		)`
	
	var exists bool
	err := r.db.QueryRow(existingQuery, verification.Date, verification.AssigneeID, verification.FamilyID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking if verification exists: %v", err)
	}

	if exists {
		query := `
			UPDATE daily_verification
			SET is_verified = $4, verified_by = $5, verified_at = $6, notes = $7, updated_at = $8
			WHERE date = $1 AND assignee_id = $2 AND family_id = $3`
		
		var verifiedBy sql.NullInt64
		var verifiedAt sql.NullTime
		
		if verification.VerifiedBy != nil {
			verifiedBy = sql.NullInt64{Int64: int64(*verification.VerifiedBy), Valid: true}
		}
		
		if verification.VerifiedAt != nil {
			verifiedAt = sql.NullTime{Time: *verification.VerifiedAt, Valid: true}
		}
		
		_, err = r.db.Exec(
			query,
			verification.Date,
			verification.AssigneeID,
			verification.FamilyID,
			verification.IsVerified,
			verifiedBy,
			verifiedAt,
			verification.Notes,
			time.Now().UTC(),
		)
		
		if err != nil {
			return fmt.Errorf("error updating verification: %v", err)
		}
	} else {
		query := `
			INSERT INTO daily_verification (
				date, assignee_id, family_id, is_verified, verified_by, verified_at, notes, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)`
		
		var verifiedBy sql.NullInt64
		var verifiedAt sql.NullTime
		
		if verification.VerifiedBy != nil {
			verifiedBy = sql.NullInt64{Int64: int64(*verification.VerifiedBy), Valid: true}
		}
		
		if verification.VerifiedAt != nil {
			verifiedAt = sql.NullTime{Time: *verification.VerifiedAt, Valid: true}
		}
		
		now := time.Now().UTC()
		_, err = r.db.Exec(
			query,
			verification.Date,
			verification.AssigneeID,
			verification.FamilyID,
			verification.IsVerified,
			verifiedBy,
			verifiedAt,
			verification.Notes,
			now,
		)
		
		if err != nil {
			return fmt.Errorf("error creating verification: %v", err)
		}
		
		verification.CreatedAt = now
		verification.UpdatedAt = now
	}
	
	return nil
}

func (r *Repository) GetDailyVerification(date time.Time, assigneeID int, familyID int) (*entities.DailyVerification, error) {
	query := `
		SELECT date, assignee_id, family_id, is_verified, verified_by, verified_at, notes, created_at, updated_at
		FROM daily_verification
		WHERE date = $1 AND assignee_id = $2 AND family_id = $3`
	
	verification := &entities.DailyVerification{}
	var verifiedBy sql.NullInt64
	var verifiedAt sql.NullTime
	
	err := r.db.QueryRow(query, date, assigneeID, familyID).Scan(
		&verification.Date,
		&verification.AssigneeID,
		&verification.FamilyID,
		&verification.IsVerified,
		&verifiedBy,
		&verifiedAt,
		&verification.Notes,
		&verification.CreatedAt,
		&verification.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return &entities.DailyVerification{
			Date:       date,
			AssigneeID: assigneeID,
			FamilyID:   familyID,
			IsVerified: false,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting verification: %v", err)
	}
	
	if verifiedBy.Valid {
		vID := int(verifiedBy.Int64)
		verification.VerifiedBy = &vID
	}
	
	if verifiedAt.Valid {
		verification.VerifiedAt = &verifiedAt.Time
	}
	
	return verification, nil
}