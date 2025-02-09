package database

import (
	"fmt"
)

func (db *PostgresDB) Init() error {
    fmt.Println("Starting database initialization...")

    if err := db.initializeSchema(); err != nil {
        return fmt.Errorf("schema initialization failed: %v", err)
    }

    // db.migrationsManager.EnableMigration("003_workspace_relationships")
    // fmt.Println("Running data migration...")
    // if err := db.migrationsManager.Run(); err != nil {
    //     return fmt.Errorf("migrations failed: %v", err)
    // }

    // UNCOMMENTING WILL DROP ALL TABLES IN DEV
    // if err := development.DropAllTables(db.DB); err != nil {
    //     log.Fatal(err)
    // }

    return nil
}

func (db *PostgresDB) initializeSchema() error {

    fmt.Println("Ensuring users table exists...")
    if err := db.createUsersTable(); err != nil {
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

    fmt.Println("Ensuring item table exists...")
    if err := db.createItemTables(); err != nil {
        return err
    }

    return nil
}