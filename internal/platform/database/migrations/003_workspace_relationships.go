package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateWorkspaceRelationships(tx *sql.Tx) error {
    // Add workspace relationship to containers
    queries := []string{
        // Add workspace_id column to container table if not exists
        `ALTER TABLE container 
         ADD COLUMN IF NOT EXISTS workspace_id INTEGER,
         ADD CONSTRAINT fk_container_workspace
             FOREIGN KEY (workspace_id)
             REFERENCES workspace(id)
             ON DELETE SET NULL;`,

        // Create index for performance
        `CREATE INDEX IF NOT EXISTS idx_container_workspace_id 
         ON container(workspace_id);`,

        // Create index for workspace search
        `CREATE INDEX IF NOT EXISTS idx_workspace_combined_search 
         ON workspace USING gin (to_tsvector('english', 
            name || ' ' || COALESCE(description, '')));`,

        // Create composite index for efficient container lookups within workspace
        `CREATE INDEX IF NOT EXISTS idx_container_workspace_user 
         ON container(workspace_id, user_id);`,

        // Create indexes for item lookups within container
        `CREATE INDEX IF NOT EXISTS idx_item_container_combined 
         ON item(container_id, created_at DESC);`,

        // Update container search index to include location
        `CREATE INDEX IF NOT EXISTS idx_container_combined_search 
         ON container USING gin (to_tsvector('english', 
            name || ' ' || COALESCE(location, '')));`,
    }

    for _, query := range queries {
        if _, err := tx.Exec(query); err != nil {
            return fmt.Errorf("failed to execute migration query: %v", err)
        }
    }

    return nil
}