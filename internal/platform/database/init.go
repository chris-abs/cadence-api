package database

import (
	"fmt"
)

func (db *PostgresDB) Init() error {
    fmt.Println("Starting database initialization...")

    if err := db.initializeSchema(); err != nil {
        return fmt.Errorf("schema initialization failed: %v", err)
    }

    db.migrationsManager.EnableMigration("002_search_indexes")
    if err := db.migrationsManager.Run(); err != nil {
        return fmt.Errorf("migrations failed: %v", err)
    }

    // UNCOMMENTING WILL DROP ALL TABLES IN DEV
    // if err := development.DropAllTables(db.DB); err != nil {
    //     log.Fatal(err)
    // }

    return nil
}

func (db *PostgresDB) initializeSchema() error {
    if err := db.createUsersTable(); err != nil {
        return err
    }

    if err := db.createWorkspaceTable(); err != nil {
        return err
    }

    if err := db.createContainerTable(); err != nil {
        return err
    }

    if err := db.createItemTables(); err != nil {
        return err
    }

    return nil
}