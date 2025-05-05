package calendar

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/calendar/entities"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) Create(event *entities.Event) error {
    query := `
        INSERT INTO calendar_event (
            title, description, location, start_time, end_time, all_day,
            source_module, source_id, created_by, assignee_id, family_id,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
        RETURNING id, created_at, updated_at`

    now := time.Now().UTC()
    
    err := r.db.QueryRow(
        query,
        event.Title,
        event.Description,
        event.Location,
        event.StartTime,
        event.EndTime,
        event.AllDay,
        event.SourceModule,
        event.SourceID,
        event.CreatedBy,
        event.AssigneeID,
        event.FamilyID,
        now,
    ).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

    if err != nil {
        return fmt.Errorf("error creating event: %v", err)
    }

    return nil
}

func (r *Repository) GetByID(id int, familyID int) (*entities.Event, error) {
    query := `
        SELECT 
            id, title, description, location, start_time, end_time, all_day,
            source_module, source_id, created_by, assignee_id, family_id,
            created_at, updated_at, is_deleted, deleted_at, deleted_by
        FROM calendar_event
        WHERE id = $1 
        AND family_id = $2 
        AND is_deleted = false`

    event := new(entities.Event)
    var description, location sql.NullString
    var sourceID sql.NullInt64
    var deletedAt sql.NullTime

    err := r.db.QueryRow(query, id, familyID).Scan(
        &event.ID,
        &event.Title,
        &description,
        &location,
        &event.StartTime,
        &event.EndTime,
        &event.AllDay,
        &event.SourceModule,
        &sourceID,
        &event.CreatedBy,
        &event.AssigneeID,
        &event.FamilyID,
        &event.CreatedAt,
        &event.UpdatedAt,
        &event.IsDeleted,
        &deletedAt,
        &event.DeletedBy,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("event not found")
    }
    if err != nil {
        return nil, fmt.Errorf("error getting event: %v", err)
    }

    if description.Valid {
        event.Description = &description.String
    }
    if location.Valid {
        event.Location = &location.String
    }
    if sourceID.Valid {
        id := int(sourceID.Int64)
        event.SourceID = &id
    }
    if deletedAt.Valid {
        event.DeletedAt = &deletedAt.Time
    }

    return event, nil
}

func (r *Repository) GetByDateRange(familyID int, params GetEventsParams) ([]*entities.Event, error) {
    query := `
        SELECT 
            id, title, description, location, start_time, end_time, all_day,
            source_module, source_id, created_by, assignee_id, family_id,
            created_at, updated_at, is_deleted, deleted_at, deleted_by
        FROM calendar_event
        WHERE family_id = $1 
        AND start_time >= $2 
        AND end_time <= $3
        AND is_deleted = false`

    args := []interface{}{familyID, params.StartTime, params.EndTime}
    argCount := 3

    if params.AssigneeID != nil {
        argCount++
        query += fmt.Sprintf(" AND assignee_id = $%d", argCount)
        args = append(args, *params.AssigneeID)
    }

    query += " ORDER BY start_time ASC"

    rows, err := r.db.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("error getting events: %v", err)
    }
    defer rows.Close()

    var events []*entities.Event
    for rows.Next() {
        event := new(entities.Event)
        var description, location sql.NullString
        var sourceID sql.NullInt64
        var deletedAt sql.NullTime

        err := rows.Scan(
            &event.ID,
            &event.Title,
            &description,
            &location,
            &event.StartTime,
            &event.EndTime,
            &event.AllDay,
            &event.SourceModule,
            &sourceID,
            &event.CreatedBy,
            &event.AssigneeID,
            &event.FamilyID,
            &event.CreatedAt,
            &event.UpdatedAt,
            &event.IsDeleted,
            &deletedAt,
            &event.DeletedBy,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning event: %v", err)
        }

        if description.Valid {
            event.Description = &description.String
        }
        if location.Valid {
            event.Location = &location.String
        }
        if sourceID.Valid {
            id := int(sourceID.Int64)
            event.SourceID = &id
        }
        if deletedAt.Valid {
            event.DeletedAt = &deletedAt.Time
        }

        events = append(events, event)
    }

    return events, nil
}

func (r *Repository) Update(event *entities.Event) error {
    query := `
        UPDATE calendar_event
        SET title = $2, description = $3, location = $4,
            start_time = $5, end_time = $6, all_day = $7,
            assignee_id = $8, updated_at = $9
        WHERE id = $1 
        AND family_id = $10 
        AND is_deleted = false`

    result, err := r.db.Exec(
        query,
        event.ID,
        event.Title,
        event.Description,
        event.Location,
        event.StartTime,
        event.EndTime,
        event.AllDay,
        event.AssigneeID,
        time.Now().UTC(),
        event.FamilyID,
    )
    if err != nil {
        return fmt.Errorf("error updating event: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking update result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("event not found")
    }

    return nil
}

func (r *Repository) Delete(id int, familyID int, deletedBy int) error {
    query := `
        UPDATE calendar_event
        SET is_deleted = true, 
            deleted_at = $3, 
            deleted_by = $4, 
            updated_at = $3
        WHERE id = $1 
        AND family_id = $2 
        AND is_deleted = false`
        
    result, err := r.db.Exec(query, id, familyID, time.Now().UTC(), deletedBy)
    if err != nil {
        return fmt.Errorf("error deleting event: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking delete result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("event not found or access denied")
    }

    return nil
}

func (r *Repository) RestoreDeleted(id int, familyID int) error {
    query := `
        UPDATE calendar_event
        SET is_deleted = false, 
            deleted_at = NULL, 
            deleted_by = NULL, 
            updated_at = $3
        WHERE id = $1 
        AND family_id = $2 
        AND is_deleted = true`
    
    result, err := r.db.Exec(query, id, familyID, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error restoring event: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking restore result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("event not found or not deleted")
    }

    return nil
}