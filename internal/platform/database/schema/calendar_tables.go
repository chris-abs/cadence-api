package schema

import (
	"database/sql"
	"fmt"
)

func InitCalendarSchema(db *sql.DB) error {
    if err := createEventTable(db); err != nil {
        return fmt.Errorf("failed to create event table: %v", err)
    }

    if err := createEventExceptionTable(db); err != nil {
        return fmt.Errorf("failed to create event exception table: %v", err)
    }

    return nil
}

func createEventTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS calendar_event (
        id SERIAL PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        description TEXT,
        location TEXT,
        start_time TIMESTAMP WITH TIME ZONE NOT NULL,
        end_time TIMESTAMP WITH TIME ZONE NOT NULL,
        all_day BOOLEAN DEFAULT FALSE,
        source_module VARCHAR(50) DEFAULT 'GENERAL',
        source_id INTEGER,
        created_by INTEGER REFERENCES profile(id),
        assignee_id INTEGER REFERENCES profile(id),
        family_id INTEGER REFERENCES family_account(id) NOT NULL,
        is_recurring BOOLEAN NOT NULL DEFAULT false,
        recurrence_type VARCHAR(50),
        recurrence_end_time TIMESTAMP WITH TIME ZONE,
        is_exception BOOLEAN NOT NULL DEFAULT false,
        parent_event_id INTEGER REFERENCES calendar_event(id),
        instance_date TIMESTAMP WITH TIME ZONE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by INTEGER REFERENCES profile(id),
        CONSTRAINT valid_recurrence_type 
        CHECK (recurrence_type IN ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY') OR recurrence_type IS NULL)
    );

    -- Core performance indexes
    CREATE INDEX IF NOT EXISTS idx_calendar_event_family_time 
        ON calendar_event(family_id, start_time) 
        WHERE is_deleted = false;
    
    CREATE INDEX IF NOT EXISTS idx_calendar_event_recurring_family 
        ON calendar_event(family_id, is_recurring, start_time) 
        WHERE is_deleted = false;
    
    CREATE INDEX IF NOT EXISTS idx_calendar_event_assignee_time 
        ON calendar_event(assignee_id, start_time) 
        WHERE is_deleted = false;
    
    CREATE INDEX IF NOT EXISTS idx_calendar_event_source_time 
        ON calendar_event(family_id, source_module, start_time) 
        WHERE is_deleted = false;
    
    CREATE INDEX IF NOT EXISTS idx_calendar_event_modified_instances 
        ON calendar_event(parent_event_id, instance_date) 
        WHERE is_exception = true AND is_deleted = false;
    
    -- Additional useful indexes
    CREATE INDEX IF NOT EXISTS idx_calendar_event_source ON calendar_event(source_module, source_id);
    CREATE INDEX IF NOT EXISTS idx_calendar_event_exception ON calendar_event(parent_event_id) WHERE parent_event_id IS NOT NULL;
    `
    
    _, err := db.Exec(query)
    return err
}

func createEventExceptionTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS calendar_event_exception (
        id SERIAL PRIMARY KEY,
        event_id INTEGER NOT NULL REFERENCES calendar_event(id) ON DELETE CASCADE,
        exception_date TIMESTAMP WITH TIME ZONE NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(event_id, exception_date)
    );

    CREATE INDEX IF NOT EXISTS idx_calendar_event_exception_date 
        ON calendar_event_exception(event_id, exception_date);
    `
    
    _, err := db.Exec(query)
    return err
}