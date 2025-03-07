package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateSoftDelete(tx *sql.Tx) error {
    tables := []string{
        "users", "family", "family_membership", "workspace", 
        "container", "item", "item_tag", "tag", "chore", "chore_instance",
    }
    
    for _, table := range tables {
        query := fmt.Sprintf(`
            ALTER TABLE %s 
            ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT false,
            ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
            ADD COLUMN IF NOT EXISTS deleted_by INTEGER REFERENCES users(id)
        `, table)
        
        if _, err := tx.Exec(query); err != nil {
            return fmt.Errorf("failed to add soft delete columns to %s: %v", table, err)
        }
        
        indexQuery := fmt.Sprintf(`
            CREATE INDEX IF NOT EXISTS idx_%s_is_deleted ON %s(is_deleted)
        `, table, table)
        
        if _, err := tx.Exec(indexQuery); err != nil {
            return fmt.Errorf("failed to create index on %s: %v", table, err)
        }
    }
    
    return nil
}