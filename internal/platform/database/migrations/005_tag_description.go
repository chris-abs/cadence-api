package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateTagDescription(tx *sql.Tx) error {
    queries := []string{
        // Add description column to tag table with a default empty string
        `ALTER TABLE tag 
         ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';`,

        // Drop existing tag search indexes that might reference the name field
        `DROP INDEX IF EXISTS idx_tag_name_pattern;`,
        `DROP INDEX IF EXISTS idx_tag_name_fts;`,

        // Recreate the pattern matching index
        `CREATE INDEX idx_tag_name_pattern 
         ON tag USING gin (name gin_trgm_ops);`,

        // Create a new full-text search index including the description
        `CREATE INDEX idx_tag_name_fts 
         ON tag USING gin (to_tsvector('english', 
            name || ' ' || COALESCE(description, '')));`,
    }

    for _, query := range queries {
        if _, err := tx.Exec(query); err != nil {
            return fmt.Errorf("failed to execute tag description migration query: %v", err)
        }
    }

    return nil
}