package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateChoreVerification(tx *sql.Tx) error {
    // Add verification columns
    verificationQuery := `
        ALTER TABLE chore_instance 
        ADD COLUMN IF NOT EXISTS verified_by INTEGER REFERENCES profile(id),
        ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP WITH TIME ZONE,
        ADD COLUMN IF NOT EXISTS rejection_reason TEXT;
    `
    if _, err := tx.Exec(verificationQuery); err != nil {
        return fmt.Errorf("failed to add verification columns to chore_instance: %v", err)
    }

    // Add verification index
    indexQuery := `
        CREATE INDEX IF NOT EXISTS idx_chore_instance_verified 
        ON chore_instance(verified_by)
        WHERE verified_by IS NOT NULL;
    `
    if _, err := tx.Exec(indexQuery); err != nil {
        return fmt.Errorf("failed to create verification index: %v", err)
    }

    // Add composite index for status + verified_by for efficient verification queries
    compositeIndexQuery := `
        CREATE INDEX IF NOT EXISTS idx_chore_instance_status_verified 
        ON chore_instance(status, verified_by)
        WHERE status IN ('completed', 'verified', 'rejected');
    `
    if _, err := tx.Exec(compositeIndexQuery); err != nil {
        return fmt.Errorf("failed to create status-verification index: %v", err)
    }

    return nil
}