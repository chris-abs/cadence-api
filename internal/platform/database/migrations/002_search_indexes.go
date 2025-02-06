package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateSearchIndexes(tx *sql.Tx) error {
    // Enable trigram extension
    if _, err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm;`); err != nil {
        return fmt.Errorf("failed to create pg_trgm extension: %v", err)
    }

    // Create search indexes for each entity
    queries := []string{
        `CREATE INDEX IF NOT EXISTS idx_tag_name_pattern ON tag USING gin (name gin_trgm_ops);`,
        `CREATE INDEX IF NOT EXISTS idx_tag_name_fts ON tag USING gin (to_tsvector('english', name));`,
        
        `CREATE INDEX IF NOT EXISTS idx_workspace_name_pattern 
         ON workspace USING gin (name gin_trgm_ops);`,
        `CREATE INDEX IF NOT EXISTS idx_workspace_name_fts 
         ON workspace USING gin (to_tsvector('english', 
            name || ' ' || COALESCE(description, '')));`,
        
        `CREATE INDEX IF NOT EXISTS idx_container_name_pattern 
         ON container USING gin (name gin_trgm_ops);`,
        `CREATE INDEX IF NOT EXISTS idx_container_name_fts 
         ON container USING gin (to_tsvector('english', name));`,
        
        `CREATE INDEX IF NOT EXISTS idx_item_name_pattern 
         ON item USING gin (name gin_trgm_ops);`,
        `CREATE INDEX IF NOT EXISTS idx_item_name_fts 
         ON item USING gin (to_tsvector('english', 
            name || ' ' || COALESCE(description, '')));`,
    }

    for _, query := range queries {
        if _, err := tx.Exec(query); err != nil {
            return fmt.Errorf("failed to create search index: %v", err)
        }
    }

    return nil
}