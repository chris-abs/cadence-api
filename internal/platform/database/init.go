package database

import (
	"fmt"
)

func (db *PostgresDB) Init() error {
    fmt.Println("Starting database initialization...")
    
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