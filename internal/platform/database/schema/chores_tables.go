package schema

import (
	"database/sql"
	"fmt"
)

func InitChoresSchema(db *sql.DB) error {
	if err := createChoreTable(db); err != nil {
		return fmt.Errorf("failed to create chore table: %v", err)
	}

	if err := createChoreInstanceTable(db); err != nil {
		return fmt.Errorf("failed to create chore instance table: %v", err)
	}

	if err := createDailyVerificationTable(db); err != nil {
		return fmt.Errorf("failed to create daily verification table: %v", err)
	}

	return nil
}

func createChoreTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS chore (
        id VARCHAR(36) PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        creator_id VARCHAR(36) REFERENCES profile(id) NOT NULL,
        assignee_id VARCHAR(36) REFERENCES profile(id) NOT NULL,
        family_id VARCHAR(36) REFERENCES family_account(id) NOT NULL,
        points INTEGER DEFAULT 0,
        occurrence_type VARCHAR(50) NOT NULL CHECK (occurrence_type IN ('daily', 'weekly', 'monthly', 'custom')),
        occurrence_data JSONB NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by VARCHAR(36) REFERENCES profile(id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_chore_family ON chore(family_id);
    CREATE INDEX IF NOT EXISTS idx_chore_assignee ON chore(assignee_id);
    CREATE INDEX IF NOT EXISTS idx_chore_creator ON chore(creator_id);
    CREATE INDEX IF NOT EXISTS idx_chore_is_deleted ON chore(is_deleted);
    CREATE INDEX IF NOT EXISTS idx_chore_family_deleted ON chore(family_id, is_deleted);
    CREATE INDEX IF NOT EXISTS idx_chore_assignee_deleted ON chore(assignee_id, is_deleted);
    CREATE INDEX IF NOT EXISTS idx_chore_occurrence_type ON chore(occurrence_type);
    `
    
    _, err := db.Exec(query)
    return err
}

func createChoreInstanceTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS chore_instance (
        id VARCHAR(36) PRIMARY KEY,
        chore_id VARCHAR(36) REFERENCES chore(id) ON DELETE CASCADE,
        assignee_id VARCHAR(36) REFERENCES profile(id) NOT NULL,
        family_id VARCHAR(36) REFERENCES family_account(id) NOT NULL,
        due_date DATE NOT NULL,
        status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'verified', 'rejected', 'missed')),
        completed_at TIMESTAMP WITH TIME ZONE,
        verified_by VARCHAR(36) REFERENCES profile(id),
        verified_at TIMESTAMP WITH TIME ZONE,
        rejection_reason TEXT,
        notes TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by VARCHAR(36) REFERENCES profile(id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_chore_instance_chore ON chore_instance(chore_id);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_family ON chore_instance(family_id);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_assignee ON chore_instance(assignee_id);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_due_date ON chore_instance(due_date);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_status ON chore_instance(status);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_is_deleted ON chore_instance(is_deleted);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_family_deleted ON chore_instance(family_id, is_deleted);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_assignee_date ON chore_instance(assignee_id, due_date);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_family_date ON chore_instance(family_id, due_date);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_assignee_status ON chore_instance(assignee_id, status, is_deleted);
    `
    
    _, err := db.Exec(query)
    return err
}

func createDailyVerificationTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS daily_verification (
        date DATE NOT NULL,
        assignee_id VARCHAR(36) REFERENCES profile(id) NOT NULL,
        family_id VARCHAR(36) REFERENCES family_account(id) NOT NULL,
        is_verified BOOLEAN NOT NULL DEFAULT false,
        verified_by VARCHAR(36) REFERENCES profile(id),
        verified_at TIMESTAMP WITH TIME ZONE,
        notes TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        
        PRIMARY KEY (date, assignee_id, family_id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_daily_verification_assignee ON daily_verification(assignee_id);
    CREATE INDEX IF NOT EXISTS idx_daily_verification_family ON daily_verification(family_id);
    CREATE INDEX IF NOT EXISTS idx_daily_verification_date ON daily_verification(date);
    CREATE INDEX IF NOT EXISTS idx_daily_verification_verified ON daily_verification(is_verified);
    `
    
    _, err := db.Exec(query)
    return err
}