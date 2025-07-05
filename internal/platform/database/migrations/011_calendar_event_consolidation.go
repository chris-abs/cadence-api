package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateCalendarConsolidation(tx *sql.Tx) error {
    // Drop existing table and recreate with correct schema
    dropTableQuery := `DROP TABLE IF EXISTS calendar_event CASCADE;`
    if _, err := tx.Exec(dropTableQuery); err != nil {
        return fmt.Errorf("error dropping existing calendar_event table: %v", err)
    }

    // Create new calendar_event table with correct schema
    createTableQuery := `
        CREATE TABLE calendar_event (
            id SERIAL PRIMARY KEY,
            title VARCHAR(255) NOT NULL,
            description TEXT,
            location TEXT,
            start_time TIMESTAMP WITH TIME ZONE NOT NULL,
            end_time TIMESTAMP WITH TIME ZONE NOT NULL,
            all_day BOOLEAN NOT NULL DEFAULT false,
            source_module VARCHAR(50) NOT NULL DEFAULT 'GENERAL',
            source_id INTEGER,
            created_by INTEGER REFERENCES profile(id) NOT NULL,
            assignee_id INTEGER REFERENCES profile(id) NOT NULL,
            family_id INTEGER REFERENCES family_account(id) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            is_deleted BOOLEAN NOT NULL DEFAULT false,
            deleted_at TIMESTAMP WITH TIME ZONE,
            deleted_by INTEGER REFERENCES profile(id)
        );

        CREATE INDEX idx_calendar_event_family ON calendar_event(family_id);
        CREATE INDEX idx_calendar_event_assignee ON calendar_event(assignee_id);
        CREATE INDEX idx_calendar_event_source ON calendar_event(source_module, source_id);
        CREATE INDEX idx_calendar_event_date ON calendar_event(start_time, end_time);
        CREATE INDEX idx_calendar_event_active_date ON calendar_event(start_time, end_time) 
            WHERE is_deleted = false;
    `

    if _, err := tx.Exec(createTableQuery); err != nil {
        return fmt.Errorf("error creating calendar_event table: %v", err)
    }

    return nil
}