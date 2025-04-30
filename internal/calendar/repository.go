package calendar

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/calendar/entities"
	"github.com/lib/pq"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) Create(event *entities.Event) error {
    query := `
        INSERT INTO event (
            title, description, start, end, all_day,
            created_by, assignee_ids, color,
            type, module_id, entity_id,
            family_id, created_at, updated_at,
            is_deleted
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, false)
        RETURNING id, created_at, updated_at`

    err := r.db.QueryRow(
        query,
        event.Title,
        event.Description,
        event.Start,
        event.End,
        event.AllDay,
        event.CreatedBy,
        pq.Array(event.AssigneeIDs),
        event.Color,
        event.Type,
        event.ModuleID,
        event.EntityID,
        event.FamilyID,
        time.Now().UTC(),
        time.Now().UTC(),
    ).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

    if err != nil {
        return fmt.Errorf("error creating event: %v", err)
    }

    return nil
}

func (r *Repository) GetByID(id int, familyID int) (*entities.Event, error) {
    query := `
        SELECT 
            id, title, description, start, end, all_day,
            created_by, assignee_ids, color,
            type, module_id, entity_id,
            family_id, created_at, updated_at
        FROM event
        WHERE id = $1 
        AND family_id = $2 
        AND is_deleted = false`

    event := new(entities.Event)
    err := r.db.QueryRow(query, id, familyID).Scan(
        &event.ID, &event.Title, &event.Description,
        &event.Start, &event.End, &event.AllDay,
        &event.CreatedBy, pq.Array(&event.AssigneeIDs), &event.Color,
        &event.Type, &event.ModuleID, &event.EntityID,
        &event.FamilyID, &event.CreatedAt, &event.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("event not found")
    }
    if err != nil {
        return nil, fmt.Errorf("error getting event: %v", err)
    }

    return event, nil
}

func (r *Repository) GetByDateRange(familyID int, params GetEventsParams) ([]*entities.Event, error) {
    query := `
        SELECT 
            id, title, description, start, end, all_day,
            created_by, assignee_ids, color,
            type, module_id, entity_id,
            family_id, created_at, updated_at
        FROM event
        WHERE family_id = $1 
        AND start >= $2 
        AND end <= $3
        AND is_deleted = false
        AND ($4::int[] IS NULL OR assignee_ids && $4)
        AND ($5::text[] IS NULL OR type = ANY($5))
        AND ($6::text[] IS NULL OR module_id = ANY($6))
        ORDER BY start ASC`

    rows, err := r.db.Query(
        query,
        familyID,
        params.Start,
        params.End,
        pq.Array(params.AssigneeIDs),
        pq.Array(params.Types),
        pq.Array(params.ModuleIDs),
    )
    if err != nil {
        return nil, fmt.Errorf("error querying events: %v", err)
    }
    defer rows.Close()

    var events []*entities.Event
    for rows.Next() {
        event := new(entities.Event)
        err := rows.Scan(
            &event.ID, &event.Title, &event.Description,
            &event.Start, &event.End, &event.AllDay,
            &event.CreatedBy, pq.Array(&event.AssigneeIDs), &event.Color,
            &event.Type, &event.ModuleID, &event.EntityID,
            &event.FamilyID, &event.CreatedAt, &event.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning event: %v", err)
        }
        events = append(events, event)
    }

    return events, nil
}

func (r *Repository) Update(event *entities.Event) error {
    query := `
        UPDATE event
        SET title = $2, description = $3,
            start = $4, end = $5, all_day = $6,
            assignee_ids = $7, color = $8,
            updated_at = $9
        WHERE id = $1 
        AND family_id = $10 
        AND is_deleted = false`

    result, err := r.db.Exec(
        query,
        event.ID,
        event.Title,
        event.Description,
        event.Start,
        event.End,
        event.AllDay,
        pq.Array(event.AssigneeIDs),
        event.Color,
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
        UPDATE event
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
        UPDATE event
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