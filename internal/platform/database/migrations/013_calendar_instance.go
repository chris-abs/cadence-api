package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateCalendarInstance(tx *sql.Tx) error {
    // Add recurrence columns to calendar_event (skip if they exist)
    recurrenceQuery := `
        ALTER TABLE calendar_event 
        ADD COLUMN IF NOT EXISTS is_recurring BOOLEAN NOT NULL DEFAULT false,
        ADD COLUMN IF NOT EXISTS recurrence_type VARCHAR(50),
        ADD COLUMN IF NOT EXISTS recurrence_end_time TIMESTAMP WITH TIME ZONE,
        ADD COLUMN IF NOT EXISTS is_exception BOOLEAN NOT NULL DEFAULT false,
        ADD COLUMN IF NOT EXISTS parent_event_id INTEGER REFERENCES calendar_event(id),
        ADD COLUMN IF NOT EXISTS instance_date TIMESTAMP WITH TIME ZONE;
    `
    if _, err := tx.Exec(recurrenceQuery); err != nil {
        return fmt.Errorf("failed to add recurrence columns to calendar_event: %v", err)
    }

    // Add constraint only if it doesn't exist
    constraintQuery := `
        DO $$ 
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM pg_constraint 
                WHERE conname = 'valid_recurrence_type'
            ) THEN
                ALTER TABLE calendar_event
                ADD CONSTRAINT valid_recurrence_type 
                CHECK (recurrence_type IN ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY') OR recurrence_type IS NULL);
            END IF;
        END $$;
    `
    if _, err := tx.Exec(constraintQuery); err != nil {
        return fmt.Errorf("failed to add recurrence type constraint: %v", err)
    }

    // Create exceptions table
    exceptionsTableQuery := `
        CREATE TABLE IF NOT EXISTS calendar_event_exception (
            id SERIAL PRIMARY KEY,
            event_id INTEGER NOT NULL REFERENCES calendar_event(id) ON DELETE CASCADE,
            exception_date TIMESTAMP WITH TIME ZONE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(event_id, exception_date)
        );
    `
    if _, err := tx.Exec(exceptionsTableQuery); err != nil {
        return fmt.Errorf("failed to create calendar_event_exception table: %v", err)
    }

    // Add indexes
    indexQueries := []string{
        `CREATE INDEX IF NOT EXISTS idx_calendar_event_recurring 
         ON calendar_event(is_recurring) 
         WHERE is_recurring = true;`,

        `CREATE INDEX IF NOT EXISTS idx_calendar_event_exception 
         ON calendar_event(parent_event_id) 
         WHERE parent_event_id IS NOT NULL;`,

        `CREATE INDEX IF NOT EXISTS idx_calendar_event_exception_date 
         ON calendar_event_exception(event_id, exception_date);`,

        `CREATE INDEX IF NOT EXISTS idx_calendar_event_dates_recurring 
         ON calendar_event(start_time, end_time, is_recurring) 
         WHERE is_deleted = false;`,

        `CREATE INDEX IF NOT EXISTS idx_calendar_event_instance_date 
         ON calendar_event(instance_date) 
         WHERE instance_date IS NOT NULL;`,
    }

    for _, query := range indexQueries {
        if _, err := tx.Exec(query); err != nil {
            return fmt.Errorf("failed to create index: %v", err)
        }
    }

    return nil
}