package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateFamilySupport(tx *sql.Tx) error {
    queries := []string{
        // Create user_role enum type
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
                "meals": {
                    "isEnabled": false,
                    "settings": {
                        "permissions": {
                            "PARENT": ["READ", "WRITE", "MANAGE"],
                            "CHILD": ["READ"]
                        }
                    }
                },
                "services": {
                    "isEnabled": false,
                    "settings": {
                        "permissions": {
                            "PARENT": ["READ", "WRITE", "MANAGE"],
                            "CHILD": ["READ"]
                        }
                    }
                },
                "chores": {
                    "isEnabled": false,
                    "settings": {
                        "permissions": {
                            "PARENT": ["READ", "WRITE", "MANAGE"],
                            "CHILD": ["READ"]
                        }
                    }
                }
            }'::jsonb,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
        );`,

        // Create family_invite table
        `CREATE TABLE IF NOT EXISTS family_invite (
            id SERIAL PRIMARY KEY,
            family_id INTEGER REFERENCES family(id) ON DELETE CASCADE,
            email VARCHAR(255) NOT NULL,
            role user_role NOT NULL,
            token VARCHAR(255) UNIQUE NOT NULL,
            expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
        );`,

        // Add family_id to core entity tables
        `ALTER TABLE workspace 
         ADD COLUMN IF NOT EXISTS family_id INTEGER REFERENCES family(id) NOT NULL;`,

        `ALTER TABLE container 
         ADD COLUMN IF NOT EXISTS family_id INTEGER REFERENCES family(id) NOT NULL;`,

        `ALTER TABLE item 
         ADD COLUMN IF NOT EXISTS family_id INTEGER REFERENCES family(id) NOT NULL;`,

        `ALTER TABLE tag 
         ADD COLUMN IF NOT EXISTS family_id INTEGER REFERENCES family(id) NOT NULL;`,

        // Add role and family_id to users
        `ALTER TABLE users 
         ADD COLUMN IF NOT EXISTS role user_role NOT NULL DEFAULT 'PARENT',
         ADD COLUMN IF NOT EXISTS family_id INTEGER REFERENCES family(id) NOT NULL;`,

        // Create indexes for performance
        `CREATE INDEX IF NOT EXISTS idx_users_family ON users(family_id);`,
        `CREATE INDEX IF NOT EXISTS idx_workspace_family ON workspace(family_id);`,
        `CREATE INDEX IF NOT EXISTS idx_container_family ON container(family_id);`,
        `CREATE INDEX IF NOT EXISTS idx_item_family ON item(family_id);`,
        `CREATE INDEX IF NOT EXISTS idx_tag_family ON tag(family_id);`,
        `CREATE INDEX IF NOT EXISTS idx_family_invite_token ON family_invite(token);`,
        `CREATE INDEX IF NOT EXISTS idx_family_invite_email ON family_invite(email);`,

        // Create indexes for foreign keys
        `CREATE INDEX IF NOT EXISTS idx_family_owner ON family(owner_id);`,
        `CREATE INDEX IF NOT EXISTS idx_workspace_user ON workspace(user_id);`,
        `CREATE INDEX IF NOT EXISTS idx_container_user ON container(user_id);`,
        `CREATE INDEX IF NOT EXISTS idx_container_workspace ON container(workspace_id);`,
        `CREATE INDEX IF NOT EXISTS idx_item_container ON item(container_id);`,
        `CREATE INDEX IF NOT EXISTS idx_item_tag ON item_tag(item_id, tag_id);`,
    }

    // Execute each query in the transaction
    for _, query := range queries {
        if _, err := tx.Exec(query); err != nil {
            return fmt.Errorf("failed to execute migration query: %v\nQuery: %s", err, query)
        }
    }

    return nil
}