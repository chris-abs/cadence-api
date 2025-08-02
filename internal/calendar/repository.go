package calendar

import (
	"database/sql"
	"fmt"
	"math/rand"
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
    event.ID = rand.Intn(900000000) + 100000000
    event.CreatedAt = time.Now().UTC()
    event.UpdatedAt = time.Now().UTC()

    query := `
        INSERT INTO calendar_event (
            id, title, description, location, start_time, end_time, all_day,
            source_module, source_id, created_by, assignee_id, family_id,
            is_recurring, recurrence_type, recurrence_end_time, instance_date,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`
    
    _, err := r.db.Exec(
        query,
        event.ID,
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
        event.IsRecurring,
        event.RecurrenceType,
        event.RecurrenceEndTime,
        event.InstanceDate,
        event.CreatedAt,
        event.UpdatedAt,
    )

    if err != nil {
        return fmt.Errorf("error creating event: %v", err)
    }

    return nil
}

func (r *Repository) scanEventWithoutProfile(scanner interface {
    Scan(dest ...interface{}) error
}) (*entities.Event, error) {
    event := new(entities.Event)
    var description, location sql.NullString
    var sourceID, parentEventID, assigneeID, deletedBy sql.NullInt64 
    var deletedAt, recurrenceEndTime, instanceDate sql.NullTime
    var recurrenceType sql.NullString

    err := scanner.Scan(
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
        &assigneeID,     
        &event.FamilyID,
        &event.IsRecurring,
        &recurrenceType,
        &recurrenceEndTime,
        &event.IsException,
        &parentEventID,
        &instanceDate,
        &event.CreatedAt,
        &event.UpdatedAt,
        &event.IsDeleted,
        &deletedAt,
        &deletedBy,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("event not found")
    }
    if err != nil {
        return nil, err
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
    if assigneeID.Valid {
        id := int(assigneeID.Int64)
        event.AssigneeID = &id
    }
    if deletedAt.Valid {
        event.DeletedAt = &deletedAt.Time
    }
    if deletedBy.Valid {
        id := int(deletedBy.Int64)
        event.DeletedBy = &id
    }
    if recurrenceType.Valid && recurrenceType.String != "" {
        rt := entities.RecurrenceType(recurrenceType.String)
        event.RecurrenceType = &rt
    }
    if recurrenceEndTime.Valid {
        event.RecurrenceEndTime = &recurrenceEndTime.Time
    }
    if parentEventID.Valid {
        id := int(parentEventID.Int64)
        event.ParentEventID = &id
    }
    if instanceDate.Valid {
        event.InstanceDate = &instanceDate.Time
    }

    return event, nil
}

func (r *Repository) batchLoadProfiles(assigneeIDs []int) (map[int]*models.Profile, error) {
    if len(assigneeIDs) == 0 {
        return make(map[int]*models.Profile), nil
    }

    uniqueIDs := make(map[int]bool)
    for _, id := range assigneeIDs {
        uniqueIDs[id] = true
    }
    
    var ids []int
    for id := range uniqueIDs {
        ids = append(ids, id)
    }

    query := `
        SELECT id, name, role, image_url, colour
        FROM profile 
        WHERE id = ANY($1)`

    rows, err := r.db.Query(query, pq.Array(ids))
    if err != nil {
        return nil, fmt.Errorf("error batch loading profiles: %v", err)
    }
    defer rows.Close()

    profiles := make(map[int]*models.Profile)
    for rows.Next() {
        profile := &models.Profile{}
        err := rows.Scan(
            &profile.ID,
            &profile.Name,
            &profile.Role,
            &profile.ImageURL,
            &profile.Colour,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning profile: %v", err)
        }
        profiles[profile.ID] = profile
    }

    return profiles, nil
}

func (r *Repository) GetByID(id int, familyID int) (*entities.Event, error) {
    query := `
        SELECT 
            e.id, e.title, e.description, e.location, e.start_time, e.end_time,
            e.all_day, e.source_module, e.source_id, e.created_by, e.assignee_id, 
            e.family_id, e.is_recurring, e.recurrence_type, e.recurrence_end_time,
            e.is_exception, e.parent_event_id, e.instance_date, e.created_at, e.updated_at,
            e.is_deleted, e.deleted_at, e.deleted_by
        FROM calendar_event e
        WHERE e.id = $1 
        AND e.family_id = $2 
        AND e.is_deleted = false`

    event, err := r.scanEventWithoutProfile(r.db.QueryRow(query, id, familyID))
    if err != nil {
        return nil, err
    }

    if event.AssigneeID != nil {
        profiles, err := r.batchLoadProfiles([]int{*event.AssigneeID})
        if err != nil {
            return nil, fmt.Errorf("error loading profile: %v", err)
        }
        if profile, exists := profiles[*event.AssigneeID]; exists {
            event.Assignee = profile
        }
    }

    return event, nil
}

func (r *Repository) GetByDateRange(familyID int, params GetEventsParams) ([]*entities.Event, error) {
    query := `
        SELECT 
            e.id, e.title, e.description, e.location, e.start_time, e.end_time,
            e.all_day, e.source_module, e.source_id, e.created_by, e.assignee_id, 
            e.family_id, e.is_recurring, e.recurrence_type, e.recurrence_end_time,
            e.is_exception, e.parent_event_id, e.instance_date, e.created_at, e.updated_at,
            e.is_deleted, e.deleted_at, e.deleted_by
        FROM calendar_event e
        WHERE e.family_id = $1 
        AND e.is_deleted = false
        AND (
            -- Single events: normal overlap check (start < queryEnd AND end > queryStart)
            (e.is_recurring = false AND e.start_time < $3 AND e.end_time > $2)
            OR 
            -- Recurring events: check if they could have instances in range
            (e.is_recurring = true 
             AND e.start_time <= $3 
             AND (e.recurrence_end_time IS NULL OR e.recurrence_end_time >= $2))
        )`

    args := []interface{}{familyID, params.StartTime, params.EndTime}
    paramCount := 3

    if len(params.AssigneeIDs) > 0 {
        paramCount++
        query += fmt.Sprintf(" AND e.assignee_id = ANY($%d)", paramCount)
        args = append(args, pq.Array(params.AssigneeIDs))
    }

    if len(params.SourceModules) > 0 {
        paramCount++
        query += fmt.Sprintf(" AND e.source_module = ANY($%d)", paramCount)
        args = append(args, pq.Array(params.SourceModules))
    }

    if params.SourceID != nil {
        paramCount++
        query += fmt.Sprintf(" AND e.source_id = $%d", paramCount)
        args = append(args, *params.SourceID)
    }

    query += " ORDER BY e.start_time ASC"

    rows, err := r.db.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("error getting events: %v", err)
    }
    defer rows.Close()

    var events []*entities.Event
    var assigneeIDs []int
    
    for rows.Next() {
        event, err := r.scanEventWithoutProfile(rows)
        if err != nil {
            return nil, fmt.Errorf("error scanning event: %v", err)
        }
        events = append(events, event)
        
        
        if event.AssigneeID != nil {
            assigneeIDs = append(assigneeIDs, *event.AssigneeID)
        }
    }

    
    if len(assigneeIDs) > 0 {
        profiles, err := r.batchLoadProfiles(assigneeIDs)
        if err != nil {
            return nil, fmt.Errorf("error batch loading profiles: %v", err)
        }
        
        
        for _, event := range events {
            if event.AssigneeID != nil {
                if profile, exists := profiles[*event.AssigneeID]; exists {
                    event.Assignee = profile
                }
            }
        }
    }

    return events, nil
}

func (r *Repository) GetExceptionsForEvents(eventIDs []int) (map[int][]time.Time, error) {
    if len(eventIDs) == 0 {
        return make(map[int][]time.Time), nil
    }

    query := `
        SELECT event_id, exception_date 
        FROM calendar_event_exception 
        WHERE event_id = ANY($1)
        ORDER BY event_id, exception_date`

    rows, err := r.db.Query(query, pq.Array(eventIDs))
    if err != nil {
        return nil, fmt.Errorf("error getting exceptions: %v", err)
    }
    defer rows.Close()

    exceptions := make(map[int][]time.Time)
    for rows.Next() {
        var eventID int
        var exceptionDate time.Time
        
        if err := rows.Scan(&eventID, &exceptionDate); err != nil {
            return nil, fmt.Errorf("error scanning exception: %v", err)
        }
        
        exceptions[eventID] = append(exceptions[eventID], exceptionDate)
    }

    return exceptions, nil
}

func (r *Repository) GetModifiedInstancesInDateRange(familyID int, startTime, endTime time.Time, parentEventIDs []int) ([]*entities.Event, error) {
    if len(parentEventIDs) == 0 {
        return []*entities.Event{}, nil
    }

    query := `
        SELECT 
            e.id, e.title, e.description, e.location, e.start_time, e.end_time,
            e.all_day, e.source_module, e.source_id, e.created_by, e.assignee_id, 
            e.family_id, e.is_recurring, e.recurrence_type, e.recurrence_end_time,
            e.is_exception, e.parent_event_id, e.instance_date, e.created_at, e.updated_at,
            e.is_deleted, e.deleted_at, e.deleted_by,
            p.id, p.name, p.role, p.image_url, p.colour
        FROM calendar_event e
        LEFT JOIN profile p ON e.assignee_id = p.id
        WHERE e.family_id = $1 
        AND e.is_exception = true
        AND e.parent_event_id = ANY($2)
        AND e.start_time < $3
        AND e.end_time > $4
        AND e.is_deleted = false
        ORDER BY e.start_time ASC`

    rows, err := r.db.Query(query, familyID, pq.Array(parentEventIDs), endTime, startTime)
    if err != nil {
        return nil, fmt.Errorf("error getting modified instances: %v", err)
    }
    defer rows.Close()

    var events []*entities.Event
    for rows.Next() {
        event, err := r.scanEvent(rows)
        if err != nil {
            return nil, fmt.Errorf("error scanning modified instance: %v", err)
        }
        events = append(events, event)
    }

    return events, nil
}

func (r *Repository) CreateModifiedInstance(event *entities.Event) error {
    query := `
        INSERT INTO calendar_event (
            title, description, location, start_time, end_time, all_day,
            source_module, source_id, created_by, assignee_id, family_id,
            is_recurring, is_exception, parent_event_id, instance_date,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $16)
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
        false, 
        true,  
        event.ParentEventID,
        event.InstanceDate,
        now,
    ).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

    if err != nil {
        return fmt.Errorf("error creating modified instance: %v", err)
    }

    return nil
}

func (r *Repository) GetModifiedInstanceByDate(parentEventID int, instanceDate time.Time) (*entities.Event, error) {
    query := `
        SELECT 
            e.id, e.title, e.description, e.location, e.start_time, e.end_time,
            e.all_day, e.source_module, e.source_id, e.created_by, e.assignee_id, 
            e.family_id, e.is_recurring, e.recurrence_type, e.recurrence_end_time,
            e.is_exception, e.parent_event_id, e.instance_date, e.created_at, e.updated_at,
            e.is_deleted, e.deleted_at, e.deleted_by,
            p.id, p.name, p.role, p.image_url, p.colour
        FROM calendar_event e
        LEFT JOIN profile p ON e.assignee_id = p.id
        WHERE e.parent_event_id = $1 
            AND e.instance_date = $2 
            AND e.is_exception = true 
            AND e.is_deleted = false
    `
    
    event, err := r.scanEvent(r.db.QueryRow(query, parentEventID, instanceDate))
    if err != nil {
        if err.Error() == "event not found" {
            return nil, fmt.Errorf("modified instance not found")
        }
        return nil, fmt.Errorf("failed to get modified instance: %v", err)
    }
    
    return event, nil
}

func (r *Repository) Update(event *entities.Event) error {
    query := `
        UPDATE calendar_event
        SET title = $2, description = $3, location = $4,
            start_time = $5, end_time = $6, all_day = $7,
            assignee_id = $8, updated_at = $9,
            recurrence_end_time = $10, instance_date = $11
        WHERE id = $1 
        AND family_id = $12 
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
        event.InstanceDate,
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

func (r *Repository) scanEvent(scanner interface {
    Scan(dest ...interface{}) error
}) (*entities.Event, error) {
    event := new(entities.Event)
    assignee := new(models.Profile)
    var description, location sql.NullString
    var sourceID, parentEventID, assigneeID, deletedBy sql.NullInt64 
    var deletedAt, recurrenceEndTime, instanceDate sql.NullTime
    var recurrenceType sql.NullString

    err := scanner.Scan(
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
        &assigneeID,     
        &event.FamilyID,
        &event.IsRecurring,
        &recurrenceType,
        &recurrenceEndTime,
        &event.IsException,
        &parentEventID,
        &instanceDate,
        &event.CreatedAt,
        &event.UpdatedAt,
        &event.IsDeleted,
        &deletedAt,
        &deletedBy,
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
        return nil, err
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
    if assigneeID.Valid {
        id := int(assigneeID.Int64)
        event.AssigneeID = &id
        event.Assignee = assignee 
    }
    if deletedAt.Valid {
        event.DeletedAt = &deletedAt.Time
    }
    if deletedBy.Valid { 
        id := int(deletedBy.Int64)
        event.DeletedBy = &id
    }
    if recurrenceType.Valid && recurrenceType.String != "" {
        rt := entities.RecurrenceType(recurrenceType.String)
        event.RecurrenceType = &rt
    }
    if recurrenceEndTime.Valid {
        event.RecurrenceEndTime = &recurrenceEndTime.Time
    }
    if parentEventID.Valid {
        id := int(parentEventID.Int64)
        event.ParentEventID = &id
    }
    if instanceDate.Valid {
        event.InstanceDate = &instanceDate.Time
    }

    return event, nil
}