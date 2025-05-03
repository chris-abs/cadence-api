package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateProfileColour(tx *sql.Tx) error {
    // Add colour column with default value
    addColumnQuery := `
        ALTER TABLE profile 
        ADD COLUMN IF NOT EXISTS colour VARCHAR(50) NOT NULL DEFAULT 'blue';
    `
    if _, err := tx.Exec(addColumnQuery); err != nil {
        return fmt.Errorf("failed to add colour column to profile: %v", err)
    }

    return nil
}