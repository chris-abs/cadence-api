package schema

import (
	"database/sql"
	"fmt"
)

func InitCalendarSchema(db *sql.DB) error {
    if err := createEventTable(db); err != nil {
        return fmt.Errorf("failed to create event table: %v", err)
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
        assignee_id INTEGER REFERENCES profile(id),
        family_id INTEGER REFERENCES family_account(id) NOT NULL,
        
        -- Recurrence fields
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
        
        -- Constraints
        CONSTRAINT valid_recurrence_type 
        CHECK (recurrence_type IN ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY') OR recurrence_type IS NULL)
    );

    -- Core indexes
    CREATE INDEX IF NOT EXISTS idx_calendar_event_family ON calendar_event(family_id);
    CREATE INDEX IF NOT EXISTS idx_calendar_event_assignee ON calendar_event(assignee_id);
    CREATE INDEX IF NOT EXISTS idx_calendar_event_source ON calendar_event(source_module, source_id);
    CREATE INDEX IF NOT EXISTS idx_calendar_event_date ON calendar_event(start_time, end_time);
    CREATE INDEX IF NOT EXISTS idx_calendar_event_active_date ON calendar_event(start_time, end_time) 
        WHERE is_deleted = false;
    
    -- Recurrence-specific indexes
    CREATE INDEX IF NOT EXISTS idx_calendar_event_recurring 
        ON calendar_event(is_recurring) 
        WHERE is_recurring = true;
    CREATE INDEX IF NOT EXISTS idx_calendar_event_exception 
        ON calendar_event(parent_event_id) 
        WHERE parent_event_id IS NOT NULL;
    CREATE INDEX IF NOT EXISTS idx_calendar_event_instance_date 
        ON calendar_event(instance_date) 
        WHERE instance_date IS NOT NULL;
    CREATE INDEX IF NOT EXISTS idx_calendar_event_dates_recurring 
        ON calendar_event(start_time, end_time, is_recurring) 
        WHERE is_deleted = false;
    `
    
    _, err := db.Exec(query)
    return err
}