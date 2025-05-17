package calendar

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/calendar/entities"
	"github.com/chrisabs/cadence/internal/models"
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
        INSERT INTO calendar_event (
            title, description, location, start_time, end_time, all_day,
            source_module, source_id, assignee_id, family_id,
            is_recurring, recurrence_type, recurrence_end_time,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $14)
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
        event.AssigneeID,
        event.FamilyID,
        event.IsRecurring,
        event.RecurrenceType,
        event.RecurrenceEndTime,
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
            e.id, e.title, e.description, e.location, e.start_time, e.end_time, 
            e.all_day, e.source_module, e.source_id, e.assignee_id, e.family_id,
            e.is_recurring, e.recurrence_type, e.recurrence_end_time,
            e.is_exception, e.parent_event_id,
            e.created_at, e.updated_at, e.is_deleted, e.deleted_at, e.deleted_by,
            p.id, p.name, p.role, p.image_url, p.colour
        FROM calendar_event e
        LEFT JOIN profile p ON e.assignee_id = p.id
        WHERE e.id = $1 AND e.family_id = $2 AND e.is_deleted = false`

    event := new(entities.Event)
    assignee := new(models.Profile)
    
    var description, location sql.NullString
    var sourceID, parentEventID sql.NullInt64
    var deletedAt, recurrenceEndTime sql.NullTime
    var recurrenceType sql.NullString

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
        &event.AssigneeID,
        &event.FamilyID,
        &event.IsRecurring,
        &recurrenceType,
        &recurrenceEndTime,
        &event.IsException,
        &parentEventID,
        &event.CreatedAt,
        &event.UpdatedAt,
        &event.IsDeleted,
        &deletedAt,
        &event.DeletedBy,
        &assignee.ID,
        &assignee.Name,
        &assignee.Role,
        &assignee.ImageURL,
        &assignee.Colour,
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
    if recurrenceType.Valid {
        event.RecurrenceType = entities.RecurrenceType(recurrenceType.String)
    }
    if recurrenceEndTime.Valid {
        event.RecurrenceEndTime = &recurrenceEndTime.Time
    }
    if parentEventID.Valid {
        id := int(parentEventID.Int64)
        event.ParentEventID = &id
    }

    event.Assignee = assignee

    return event, nil
}

func (r *Repository) GetByDateRange(familyID int, params GetEventsParams) ([]*entities.Event, int, error) {
    // First get total count for pagination
    countQuery := `
        SELECT COUNT(*) FROM (
            SELECT 1 FROM calendar_event e
            WHERE e.family_id = $1 
            AND e.start_time < $2
            AND e.end_time > $3
            AND e.is_deleted = false
            AND NOT EXISTS (
                -- Exclude instances that have been modified (they'll be included separately)
                SELECT 1 FROM calendar_event modified
                WHERE modified.parent_event_id = e.id
                AND modified.start_time = e.start_time
                AND modified.is_deleted = false
            )
            UNION ALL
            -- Include modified instances
            SELECT 1 FROM calendar_event e
            WHERE e.family_id = $1
            AND e.start_time < $2
            AND e.end_time > $3
            AND e.is_deleted = false
            AND e.is_exception = true
        ) as combined`

    var total int
    err := r.db.QueryRow(countQuery, familyID, params.EndTime, params.StartTime).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("error getting count: %v", err)
    }

    // Then get paginated results
    query := `
        WITH combined_events AS (
            -- Regular and recurring events
            SELECT e.*, NULL as exception_date
            FROM calendar_event e
            WHERE e.family_id = $1 
            AND e.start_time < $2
            AND e.end_time > $3
            AND e.is_deleted = false
            AND NOT EXISTS (
                SELECT 1 FROM calendar_event modified
                WHERE modified.parent_event_id = e.id
                AND modified.start_time = e.start_time
                AND modified.is_deleted = false
            )
            UNION ALL
            -- Modified instances
            SELECT e.*, NULL as exception_date
            FROM calendar_event e
            WHERE e.family_id = $1
            AND e.start_time < $2
            AND e.end_time > $3
            AND e.is_deleted = false
            AND e.is_exception = true
        )
        SELECT 
            e.id, e.title, e.description, e.location, e.start_time, e.end_time,
            e.all_day, e.source_module, e.source_id, e.assignee_id, e.family_id,
            e.is_recurring, e.recurrence_type, e.recurrence_end_time,
            e.is_exception, e.parent_event_id,
            e.created_at, e.updated_at, e.is_deleted, e.deleted_at, e.deleted_by,
            p.id, p.name, p.role, p.image_url, p.colour,
            ce.exception_date
        FROM combined_events e
        LEFT JOIN profile p ON e.assignee_id = p.id
        LEFT JOIN calendar_event_exception ce ON e.id = ce.event_id 
            AND ce.exception_date BETWEEN $3 AND $2`

    args := []interface{}{familyID, params.EndTime, params.StartTime}
    paramCount := 3

    if len(params.AssigneeIDs) > 0 {
        paramCount++
        query += fmt.Sprintf(" AND e.assignee_id = ANY($%d)", paramCount)
        args = append(args, pq.Array(params.AssigneeIDs))
    }

    if len(params.ModuleIDs) > 0 {
        paramCount++
        query += fmt.Sprintf(" AND e.source_module = ANY($%d)", paramCount)
        args = append(args, pq.Array(params.ModuleIDs))
    }

    if params.SourceID != nil {
        paramCount++
        query += fmt.Sprintf(" AND e.source_id = $%d", paramCount)
        args = append(args, *params.SourceID)
    }

    query += " ORDER BY e.start_time ASC"
    
    // Add pagination
    if params.Limit > 0 {
        paramCount++
        query += fmt.Sprintf(" LIMIT $%d", paramCount)
        args = append(args, params.Limit)

        if params.Offset > 0 {
            paramCount++
            query += fmt.Sprintf(" OFFSET $%d", paramCount)
            args = append(args, params.Offset)
        }
    }

    rows, err := r.db.Query(query, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("error getting events: %v", err)
    }
    defer rows.Close()

    var events []*entities.Event
    for rows.Next() {
        event := new(entities.Event)
        assignee := new(models.Profile)
        var description, location sql.NullString
        var sourceID, parentEventID sql.NullInt64
        var deletedAt, recurrenceEndTime, exceptionDate sql.NullTime
        var recurrenceType sql.NullString

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
            &event.AssigneeID,
            &event.FamilyID,
            &event.IsRecurring,
            &recurrenceType,
            &recurrenceEndTime,
            &event.IsException,
            &parentEventID,
            &event.CreatedAt,
            &event.UpdatedAt,
            &event.IsDeleted,
            &deletedAt,
            &event.DeletedBy,
            &assignee.ID,
            &assignee.Name,
            &assignee.Role,
            &assignee.ImageURL,
            &assignee.Colour,
            &exceptionDate,
        )
        if err != nil {
            return nil, 0, fmt.Errorf("error scanning event: %v", err)
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
        if recurrenceType.Valid {
            event.RecurrenceType = entities.RecurrenceType(recurrenceType.String)
        }
        if recurrenceEndTime.Valid {
            event.RecurrenceEndTime = &recurrenceEndTime.Time
        }
        if parentEventID.Valid {
            id := int(parentEventID.Int64)
            event.ParentEventID = &id
        }

        event.Assignee = assignee
        events = append(events, event)
    }

    return events, total, nil
}

func (r *Repository) Update(event *entities.Event) error {
    query := `
        UPDATE calendar_event
        SET title = $2, description = $3, location = $4,
            start_time = $5, end_time = $6, all_day = $7,
            assignee_id = $8, updated_at = $9,
            recurrence_end_time = $10
        WHERE id = $1 
        AND family_id = $11 
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
        event.RecurrenceEndTime,
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

func (r *Repository) CreateModifiedInstance(event *entities.Event) error {
    query := `
        INSERT INTO calendar_event (
            title, description, location, start_time, end_time, all_day,
            source_module, source_id, assignee_id, family_id,
            is_recurring, is_exception, parent_event_id,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $14)
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
        event.AssigneeID,
        event.FamilyID,
        false, // is_recurring
        true,  // is_exception
        event.ParentEventID,
        now,
    ).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

    if err != nil {
        return fmt.Errorf("error creating modified instance: %v", err)
    }

    return nil
}

func (r *Repository) CreateRecurrenceException(eventID int, date time.Time) error {
    query := `
        INSERT INTO calendar_event_exception (event_id, exception_date)
        VALUES ($1, $2)
        ON CONFLICT (event_id, exception_date) DO NOTHING`

    _, err := r.db.Exec(query, eventID, date)
    if err != nil {
        return fmt.Errorf("error creating recurrence exception: %v", err)
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
        WHERE (id = $1 OR parent_event_id = $1)
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
        WHERE (id = $1 OR parent_event_id = $1)
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