package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateCalendarInstanceDate(tx *sql.Tx) error {
    // Add instance_date column to calendar_event
    instanceDateQuery := `
        ALTER TABLE calendar_event 
        ADD COLUMN IF NOT EXISTS instance_date TIMESTAMP WITH TIME ZONE;
    `
    if _, err := tx.Exec(instanceDateQuery); err != nil {
        return fmt.Errorf("failed to add instance_date column to calendar_event: %v", err)
    }

    // Add index for efficient instance date queries
    indexQuery := `
        CREATE INDEX IF NOT EXISTS idx_calendar_event_instance_date 
        ON calendar_event(instance_date) 
        WHERE instance_date IS NOT NULL;
    `
    if _, err := tx.Exec(indexQuery); err != nil {
        return fmt.Errorf("failed to create instance_date index: %v", err)
    }

    // Add comment explaining the field usage
    commentQuery := `
        COMMENT ON COLUMN calendar_event.instance_date IS 
        'For recurring instances: the specific occurrence date. For exceptions: the original date that was modified. NULL for regular events.';
    `
    if _, err := tx.Exec(commentQuery); err != nil {
        // Comment failures are non-critical
        fmt.Printf("Warning: failed to add column comment: %v\n", err)
    }

    return nil
}