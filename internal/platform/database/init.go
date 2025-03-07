package database

import (
	"fmt"
	"os"

	"github.com/chrisabs/cadence/internal/platform/database/development"
	"github.com/chrisabs/cadence/internal/platform/database/schema"
)

func (db *PostgresDB) Init() error {
	fmt.Println("Starting database initialization...")

	// db.migrationsManager = migrations.NewManager(db.DB)
	// db.migrationsManager.EnableMigration("007_soft_delete")

	// fmt.Println("Running soft delete migration...")
	// if err := db.migrationsManager.Run(); err != nil {
   	// 	 return fmt.Errorf("migrations failed: %v", err)
	// }

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
	fmt.Println("Initializing core schema...")
	if err := schema.InitCoreSchema(db.DB); err != nil {
		return fmt.Errorf("core schema initialization failed: %v", err)
	}

	fmt.Println("Initializing module schemas...")
	
	if err := schema.InitStorageSchema(db.DB); err != nil {
		return fmt.Errorf("storage module schema initialization failed: %v", err)
	}
	
	if err := schema.InitChoresSchema(db.DB); err != nil {
		return fmt.Errorf("chores module schema initialization failed: %v", err)
	}
	
	if err := schema.InitMealsSchema(db.DB); err != nil {
		return fmt.Errorf("meals module schema initialization failed: %v", err)
	}
	
	if err := schema.InitServicesSchema(db.DB); err != nil {
		return fmt.Errorf("services module schema initialization failed: %v", err)
	}

	return nil
}

func (db *PostgresDB) createEnums() error {
	query := `DO $$ 
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
            CREATE TYPE user_role AS ENUM ('PARENT', 'CHILD');
        END IF;
    END $$;`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create enum: %v", err)
	}

	return nil
}