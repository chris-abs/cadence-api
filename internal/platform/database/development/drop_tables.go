package development

import (
	"database/sql"
	"fmt"
)

func DropAllTables(db *sql.DB) error {
    existingTables, err := getCurrentTables(db)
    if err != nil {
        return err
    }

    fmt.Println("Current tables before dropping:")
    for _, table := range existingTables {
        fmt.Printf("- %s\n", table)
    }

    dropServiceModuleTables := `
        DROP TABLE IF EXISTS service_payment CASCADE;
        DROP TABLE IF EXISTS service CASCADE;
    `

    dropMealsModuleTables := `
        DROP TABLE IF EXISTS shopping_list_item CASCADE;
        DROP TABLE IF EXISTS shopping_list CASCADE;
        DROP TABLE IF EXISTS meal_plan_assignee CASCADE;
        DROP TABLE IF EXISTS meal_plan CASCADE;
        DROP TABLE IF EXISTS recipe CASCADE;
    `

    dropChoresModuleTables := `
        DROP TABLE IF EXISTS chore_instance CASCADE;
        DROP TABLE IF EXISTS chore CASCADE;
    `

    dropStorageModuleTables := `
        DROP TABLE IF EXISTS item_tag CASCADE;
        DROP TABLE IF EXISTS item_image CASCADE;
        DROP TABLE IF EXISTS item CASCADE;
        DROP TABLE IF EXISTS tag CASCADE;
        DROP TABLE IF EXISTS container CASCADE;
        DROP TABLE IF EXISTS workspace CASCADE;
    `

    dropCoreTables := `
        DROP TABLE IF EXISTS notification CASCADE;
        DROP TABLE IF EXISTS calendar_event CASCADE;
        DROP TABLE IF EXISTS family_invite CASCADE;
        DROP TABLE IF EXISTS family_membership CASCADE;
        DROP TABLE IF EXISTS family CASCADE;
        DROP TABLE IF EXISTS users CASCADE;
    `

    fmt.Println("Dropping service module tables...")
    if _, err := db.Exec(dropServiceModuleTables); err != nil {
        return fmt.Errorf("error dropping service module tables: %v", err)
    }

    fmt.Println("Dropping meals module tables...")
    if _, err := db.Exec(dropMealsModuleTables); err != nil {
        return fmt.Errorf("error dropping meals module tables: %v", err)
    }

    fmt.Println("Dropping chores module tables...")
    if _, err := db.Exec(dropChoresModuleTables); err != nil {
        return fmt.Errorf("error dropping chores module tables: %v", err)
    }

    fmt.Println("Dropping storage module tables...")
    if _, err := db.Exec(dropStorageModuleTables); err != nil {
        return fmt.Errorf("error dropping storage module tables: %v", err)
    }

    fmt.Println("Dropping core tables...")
    if _, err := db.Exec(dropCoreTables); err != nil {
        return fmt.Errorf("error dropping core tables: %v", err)
    }

    remainingTables, err := getCurrentTables(db)
    if err != nil {
        return err
    }

    if len(remainingTables) > 0 {
        fmt.Println("Dropping remaining tables:")
        for _, table := range remainingTables {
            fmt.Printf("- %s\n", table)
            if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", table)); err != nil {
                return fmt.Errorf("error dropping table %s: %v", table, err)
            }
        }
    }

    dropEnums := `
    DO $$ 
    DECLARE
        enum_type text;
    BEGIN
        FOR enum_type IN (SELECT t.typname FROM pg_type t JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace WHERE t.typtype = 'e' AND n.nspname = 'public')
        LOOP
            EXECUTE 'DROP TYPE IF EXISTS ' || enum_type || ' CASCADE;';
        END LOOP;
    END $$;
    `
    
    fmt.Println("Dropping enum types...")
    if _, err := db.Exec(dropEnums); err != nil {
        return fmt.Errorf("error dropping enum types: %v", err)
    }

    finalRemainingTables, err := getCurrentTables(db)
    if err != nil {
        return err
    }

    fmt.Println("Remaining tables after dropping:")
    for _, table := range finalRemainingTables {
        fmt.Printf("- %s\n", table)
    }

    return nil
}

func getCurrentTables(db *sql.DB) ([]string, error) {
    query := `
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema = 'public'
        AND table_type = 'BASE TABLE';
    `

    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("error checking existing tables: %v", err)
    }
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var tableName string
        if err := rows.Scan(&tableName); err != nil {
            return nil, fmt.Errorf("error scanning table name: %v", err)
        }
        tables = append(tables, tableName)
    }

    return tables, nil
}