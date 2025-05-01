package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateCalendarSupport(tx *sql.Tx) error {
    // Create event table
    eventQuery := `
        CREATE TABLE IF NOT EXISTS event (
            id SERIAL PRIMARY KEY,
            title VARCHAR(255) NOT NULL,
            description TEXT,
            start TIMESTAMP WITH TIME ZONE NOT NULL,
            end TIMESTAMP WITH TIME ZONE NOT NULL,
            all_day BOOLEAN NOT NULL DEFAULT false,
            created_by INTEGER REFERENCES profile(id) NOT NULL,
            assignee_ids INTEGER[] NOT NULL DEFAULT '{}',
            color VARCHAR(7),
            type VARCHAR(50) NOT NULL DEFAULT 'GENERAL',
            module_id VARCHAR(50),
            entity_id INTEGER,
            family_id INTEGER REFERENCES family_account(id) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            is_deleted BOOLEAN NOT NULL DEFAULT false,
            deleted_at TIMESTAMP WITH TIME ZONE,
            deleted_by INTEGER REFERENCES profile(id)
        );
    `
    if _, err := tx.Exec(eventQuery); err != nil {
        return fmt.Errorf("failed to create event table: %v", err)
    }

    // Create indexes
    indexQueries := []string{
        `CREATE INDEX IF NOT EXISTS idx_event_family ON event(family_id);`,
        `CREATE INDEX IF NOT EXISTS idx_event_creator ON event(created_by);`,
        `CREATE INDEX IF NOT EXISTS idx_event_date_range ON event(start, end);`,
        `CREATE INDEX IF NOT EXISTS idx_event_type ON event(type);`,
        `CREATE INDEX IF NOT EXISTS idx_event_module ON event(module_id) WHERE module_id IS NOT NULL;`,
        `CREATE INDEX IF NOT EXISTS idx_event_is_deleted ON event(is_deleted);`,
        `CREATE INDEX IF NOT EXISTS idx_event_family_deleted ON event(family_id, is_deleted);`,
        // Composite index for date range queries within a family
        `CREATE INDEX IF NOT EXISTS idx_event_family_dates ON event(family_id, start, end) WHERE is_deleted = false;`,
    }

    for _, query := range indexQueries {
        if _, err := tx.Exec(query); err != nil {
            return fmt.Errorf("failed to create index: %v", err)
        }
    }

    return nil
}