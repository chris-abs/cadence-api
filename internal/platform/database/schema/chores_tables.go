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

	return nil
}

func createChoreTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS chore (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        creator_id INTEGER REFERENCES users(id) NOT NULL,
        assignee_id INTEGER REFERENCES users(id) NOT NULL,
        family_id INTEGER REFERENCES family(id) NOT NULL,
        points INTEGER DEFAULT 0,
        occurrence_type VARCHAR(50) NOT NULL,
        occurrence_data JSONB NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_chore_family ON chore(family_id);
    CREATE INDEX IF NOT EXISTS idx_chore_assignee ON chore(assignee_id);
    CREATE INDEX IF NOT EXISTS idx_chore_creator ON chore(creator_id);
    `
    
    _, err := db.Exec(query)
    return err
}

func createChoreInstanceTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS chore_instance (
        id SERIAL PRIMARY KEY,
        chore_id INTEGER REFERENCES chore(id) ON DELETE CASCADE,
        assignee_id INTEGER REFERENCES users(id) NOT NULL,
        family_id INTEGER REFERENCES family(id) NOT NULL,
        due_date DATE NOT NULL,
        status VARCHAR(50) NOT NULL DEFAULT 'pending',
        completed_at TIMESTAMP WITH TIME ZONE,
        notes TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_chore_instance_chore ON chore_instance(chore_id);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_family ON chore_instance(family_id);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_assignee ON chore_instance(assignee_id);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_due_date ON chore_instance(due_date);
    CREATE INDEX IF NOT EXISTS idx_chore_instance_status ON chore_instance(status);
    `
    
    _, err := db.Exec(query)
    return err
}