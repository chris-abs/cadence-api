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

	fmt.Println("Ensuring family membership table exists...")
	if err := db.createFamilyMembershipTable(); err != nil {
		return err
	}

	fmt.Println("Ensuring family invite table exists...")
	if err := db.createFamilyInviteTable(); err != nil {
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