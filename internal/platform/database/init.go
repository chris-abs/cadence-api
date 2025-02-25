package database

import (
	"fmt"
	"os"

	"github.com/chrisabs/storage/internal/platform/database/development"
)

func (db *PostgresDB) Init() error {
	fmt.Println("Starting database initialization...")

	if os.Getenv("DROP_TABLES") == "true" {
		fmt.Println("DROP_TABLES environment variable is set to true. Dropping all tables...")
		if err := development.DropAllTables(db.DB); err != nil {
			return fmt.Errorf("failed to drop tables: %v", err)
		}
		fmt.Println("Tables dropped successfully.")
	}

	if err := db.createEnums(); err != nil {
		return fmt.Errorf("enum initialization failed: %v", err)
	}

	if err := db.initializeSchema(); err != nil {
		return fmt.Errorf("schema initialization failed: %v", err)
	}

    if err := db.addForeignKeyConstraints(); err != nil {
        return fmt.Errorf("failed to add foreign key constraints: %v", err)
    }

	return nil
}

func (db *PostgresDB) initializeSchema() error {
	fmt.Println("Ensuring users table exists...")
	if err := db.createUsersTable(); err != nil {
		return err
	}

	fmt.Println("Ensuring family tables exist...")
	if err := db.createFamilyTables(); err != nil {
		return err
	}

	fmt.Println("Ensuring workspace table exists...")
	if err := db.createWorkspaceTable(); err != nil {
		return err
	}

	fmt.Println("Ensuring container table exists...")
	if err := db.createContainerTable(); err != nil {
		return err
	}

	fmt.Println("Ensuring item tables exist...")
	if err := db.createItemTables(); err != nil {
		return err
	}

	return nil
}

func (db *PostgresDB) addForeignKeyConstraints() error {
    fmt.Println("Adding foreign key constraints...")
    
    query := `
        DO $$
        BEGIN
            -- Add users.family_id foreign key if it doesn't exist
            IF NOT EXISTS (
                SELECT 1 FROM information_schema.table_constraints
                WHERE constraint_name = 'fk_users_family'
            ) THEN
                ALTER TABLE users ADD CONSTRAINT fk_users_family FOREIGN KEY (family_id) REFERENCES family(id);
            END IF;
            
            -- Add family.owner_id foreign key if it doesn't exist
            IF NOT EXISTS (
                SELECT 1 FROM information_schema.table_constraints
                WHERE constraint_name = 'fk_family_owner'
            ) THEN
                ALTER TABLE family ADD CONSTRAINT fk_family_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE RESTRICT;
            END IF;
        END
        $$;
    `
    
    _, err := db.Exec(query)
    if err != nil {
        return fmt.Errorf("failed to add foreign key constraints: %v", err)
    }
    fmt.Println("Foreign key constraints check completed.")
    return nil
}
