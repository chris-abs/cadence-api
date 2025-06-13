package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateCalendarIndexes(tx *sql.Tx) error {
    // Add performance indexes for calendar queries
    indexQueries := []string{
        // Core calendar query performance - for GetByDateRange filtering
        `CREATE INDEX IF NOT EXISTS idx_calendar_event_family_time 
         ON calendar_event(family_id, start_time) 
         WHERE is_deleted = false;`,

        // Recurring events performance - for recurring event queries  
        `CREATE INDEX IF NOT EXISTS idx_calendar_event_recurring_family 
         ON calendar_event(family_id, is_recurring, start_time) 
         WHERE is_deleted = false;`,

        // Assignee filtering performance - for assigneeIds parameter
        `CREATE INDEX IF NOT EXISTS idx_calendar_event_assignee_time 
         ON calendar_event(assignee_id, start_time) 
         WHERE is_deleted = false;`,

        // Source module filtering performance - for sourceModules parameter
        `CREATE INDEX IF NOT EXISTS idx_calendar_event_source_time 
         ON calendar_event(family_id, source_module, start_time) 
         WHERE is_deleted = false;`,

        // Modified instances performance - for exception handling
        `CREATE INDEX IF NOT EXISTS idx_calendar_event_modified_instances 
         ON calendar_event(parent_event_id, instance_date) 
         WHERE is_exception = true AND is_deleted = false;`,
    }

    for i, query := range indexQueries {
        if _, err := tx.Exec(query); err != nil {
            return fmt.Errorf("failed to create performance index %d: %v", i+1, err)
        }
    }

    return nil
}