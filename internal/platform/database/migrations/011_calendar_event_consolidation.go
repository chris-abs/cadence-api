package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateCalendarConsolidation(tx *sql.Tx) error {
    // First check if old 'event' table exists
    checkOldTableQuery := `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_name = 'event'
        );`
    
    var oldTableExists bool
    if err := tx.QueryRow(checkOldTableQuery).Scan(&oldTableExists); err != nil {
        return fmt.Errorf("error checking old table existence: %v", err)
    }

    // Check if new 'calendar_event' table exists
    checkNewTableQuery := `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_name = 'calendar_event'
        );`
    
    var newTableExists bool
    if err := tx.QueryRow(checkNewTableQuery).Scan(&newTableExists); err != nil {
        return fmt.Errorf("error checking new table existence: %v", err)
    }

    // Create new calendar_event table if it doesn't exist
    if !newTableExists {
        createTableQuery := `
            CREATE TABLE calendar_event (
                id SERIAL PRIMARY KEY,
                title VARCHAR(255) NOT NULL,
                description TEXT,
                location TEXT,                     -- Moved up with other descriptive fields
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

            -- Indexes remain the same
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
    } else {
        // If the new table exists but needs the location column
        checkLocationColumnQuery := `
            SELECT EXISTS (
                SELECT FROM information_schema.columns 
                WHERE table_name = 'calendar_event' AND column_name = 'location'
            );`

        var locationColumnExists bool
        if err := tx.QueryRow(checkLocationColumnQuery).Scan(&locationColumnExists); err != nil {
            return fmt.Errorf("error checking location column existence: %v", err)
        }

        if !locationColumnExists {
            // Add location column after description
            alterTableQuery := `
                ALTER TABLE calendar_event 
                ADD COLUMN location TEXT
                -- Position column after description
                AFTER description;`

            if _, err := tx.Exec(alterTableQuery); err != nil {
                // If AFTER clause not supported (depends on PostgreSQL version), try without positioning
                alterTableQuery = `
                    ALTER TABLE calendar_event 
                    ADD COLUMN location TEXT;`
                if _, err := tx.Exec(alterTableQuery); err != nil {
                    return fmt.Errorf("error adding location column: %v", err)
                }
            }
        }
    }

    // If old table exists, migrate data to new table
    if oldTableExists {
        // Migrate data from old table to new table
        migrateDataQuery := `
            INSERT INTO calendar_event (
                title, description, start_time, end_time, all_day,
                created_by, assignee_id, type, source_module, source_id,
                family_id, created_at, updated_at, is_deleted,
                deleted_at, deleted_by
            )
            SELECT 
                title, description, "start", "end", all_day,
                created_by, 
                CASE 
                    WHEN array_length(assignee_ids, 1) > 0 THEN assignee_ids[1]
                    ELSE created_by
                END as assignee_id,
                type, COALESCE(module_id, 'GENERAL'), entity_id,
                family_id, created_at, updated_at, is_deleted,
                deleted_at, deleted_by
            FROM event
            WHERE NOT EXISTS (
                SELECT 1 FROM calendar_event ce 
                WHERE ce.source_module = event.module_id 
                AND ce.source_id = event.entity_id
            );
        `

        if _, err := tx.Exec(migrateDataQuery); err != nil {
            return fmt.Errorf("error migrating data: %v", err)
        }

        // Drop old table
        dropOldTableQuery := `DROP TABLE IF EXISTS event;`
        if _, err := tx.Exec(dropOldTableQuery); err != nil {
            return fmt.Errorf("error dropping old table: %v", err)
        }
    }

    // If the new table exists but needs the location column
    if newTableExists {
        // Check if location column exists
        checkLocationColumnQuery := `
            SELECT EXISTS (
                SELECT FROM information_schema.columns 
                WHERE table_name = 'calendar_event' AND column_name = 'location'
            );`

        var locationColumnExists bool
        if err := tx.QueryRow(checkLocationColumnQuery).Scan(&locationColumnExists); err != nil {
            return fmt.Errorf("error checking location column existence: %v", err)
        }

        if !locationColumnExists {
            // Add location column
            alterTableQuery := `
                ALTER TABLE calendar_event 
                ADD COLUMN location TEXT;`

            if _, err := tx.Exec(alterTableQuery); err != nil {
                return fmt.Errorf("error adding location column: %v", err)
            }
        }
    }

    return nil
}