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

    dropQuery := `
        DROP TABLE IF EXISTS item_tag CASCADE;
        DROP TABLE IF EXISTS tag CASCADE;
        DROP TABLE IF EXISTS item_image CASCADE;
        DROP TABLE IF EXISTS item CASCADE;
        DROP TABLE IF EXISTS container CASCADE;
        DROP TABLE IF EXISTS workspace CASCADE;
        DROP TABLE IF EXISTS users CASCADE;
        DROP TABLE IF EXISTS family_invite CASCADE;
        DROP TABLE IF EXISTS family CASCADE;
    `

    fmt.Println("Executing drop tables...")
    if _, err := db.Exec(dropQuery); err != nil {
        return fmt.Errorf("error dropping tables: %v", err)
    }

    remainingTables, err := getCurrentTables(db)
    if err != nil {
        return err
    }

    fmt.Println("Remaining tables after dropping:")
    for _, table := range remainingTables {
        fmt.Printf("- %s\n", table)
    }

    return nil
}

func getCurrentTables(db *sql.DB) ([]string, error) {
    query := `
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema = 'public';
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