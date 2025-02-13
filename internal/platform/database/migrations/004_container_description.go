package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateContainerDescription(tx *sql.Tx) error {
    queries := []string{
        // Add description column to container table
        `ALTER TABLE container 
         ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';`,

        // Update the existing combined search index to include description
        `DROP INDEX IF EXISTS idx_container_combined_search;`,
        
        `CREATE INDEX idx_container_combined_search 
         ON container USING gin (to_tsvector('english', 
            name || ' ' || 
            COALESCE(description, '') || ' ' || 
            COALESCE(location, '')));`,
    }

    for _, query := range queries {
        if _, err := tx.Exec(query); err != nil {
            return fmt.Errorf("failed to execute container description migration query: %v", err)
        }
    }

    return nil
}