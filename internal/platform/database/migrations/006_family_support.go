package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateFamilySupport(tx *sql.Tx) error {
    queries := []string{
        // Create user_role enum
        `DO $$ BEGIN
            CREATE TYPE user_role AS ENUM ('PARENT', 'CHILD');
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;`,

        // Create families table
        `CREATE TABLE IF NOT EXISTS family (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            owner_id INTEGER REFERENCES users(id),
            module_permissions JSONB NOT NULL DEFAULT '{
                "storage": {
                    "isEnabled": true,
                    "settings": {
                        "permissions": {
                            "PARENT": ["READ", "WRITE", "MANAGE"],
                            "CHILD": ["READ"]
                        }
                    }
                },
                "meals": {"isEnabled": false},
                "services": {"isEnabled": false},
                "chores": {"isEnabled": false}
            }'::jsonb,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL
        );`,

        // Add family-related columns to users - removed SET NULL
        `ALTER TABLE users 
         ADD COLUMN IF NOT EXISTS role user_role NOT NULL DEFAULT 'PARENT',
         ADD COLUMN IF NOT EXISTS family_id INTEGER NOT NULL REFERENCES family(id);`,

        // Create family_invite table
        `CREATE TABLE IF NOT EXISTS family_invite (
            id SERIAL PRIMARY KEY,
            family_id INTEGER REFERENCES family(id) ON DELETE CASCADE,
            email VARCHAR(255) NOT NULL,
            role user_role NOT NULL,
            token VARCHAR(255) UNIQUE NOT NULL,
            expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL
        );`,

        // Add indexes
        `CREATE INDEX IF NOT EXISTS idx_users_family ON users(family_id);`,
        `CREATE INDEX IF NOT EXISTS idx_family_invite_token ON family_invite(token);`,
        `CREATE INDEX IF NOT EXISTS idx_family_invite_email ON family_invite(email);`,
    }

    for _, query := range queries {
        if _, err := tx.Exec(query); err != nil {
            return fmt.Errorf("failed to execute family support migration query: %v", err)
        }
    }

    return nil
}