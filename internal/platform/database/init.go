package database

import (
	"fmt"
)

func (db *PostgresDB) Init() error {
    fmt.Println("Starting database initialization...")

    if err := db.InitializeTables(); err != nil {
        return fmt.Errorf("schema initialization failed: %v", err)
    }

    // db.migrationsManager.EnableMigration("006_family_support")
    // fmt.Println("Running data migration...")
    // if err := db.migrationsManager.Run(); err != nil {
    //     return fmt.Errorf("migrations failed: %v", err)
    // }

    return nil
}