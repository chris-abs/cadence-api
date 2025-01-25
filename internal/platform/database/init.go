package database

import "fmt"

func (db *PostgresDB) Init() error {
    fmt.Println("Starting database initialization...")

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

    fmt.Println("Ensuring item tables exist...")
    if err := db.createItemTables(); err != nil {
        return err
    }

    // Uncomment to run migrations
    /*
    fmt.Println("Running item images migration...")
    if err := db.MigrateItemImages(); err != nil {
        return err
    }
    */

    return nil
}

// For development use only
/* 
func (db *PostgresDB) DropAllTables() error {
	 checkQuery := `
	 SELECT table_name
	 FROM information_schema.tables
	 WHERE table_schema = 'public';`

 rows, err := db.Query(checkQuery)
 if err != nil {
	 return fmt.Errorf("error checking existing tables: %v", err)
 }
 defer rows.Close()

 fmt.Println("Current tables before dropping:")
 for rows.Next() {
	 var tableName string
	 rows.Scan(&tableName)
	 fmt.Printf("- %s\n", tableName)
 }

 // Drop tables
 dropQuery := `
	 DROP TABLE IF EXISTS item_tag CASCADE;
	 DROP TABLE IF EXISTS tag CASCADE;
	 DROP TABLE IF EXISTS item CASCADE;
	 DROP TABLE IF EXISTS container CASCADE;
	 DROP TABLE IF EXISTS workspace CASCADE;
	 DROP TABLE IF EXISTS users CASCADE;
 `

 fmt.Println("Executing drop tables...")
 _, err = db.Exec(dropQuery)
 if err != nil {
	 return fmt.Errorf("error dropping tables: %v", err)
 }

 // Check remaining tables
 rows, err = db.Query(checkQuery)
 if err != nil {
	 return fmt.Errorf("error checking tables after drop: %v", err)
 }
 defer rows.Close()

 fmt.Println("Remaining tables after dropping:")
 for rows.Next() {
	 var tableName string
	 rows.Scan(&tableName)
	 fmt.Printf("- %s\n", tableName)
  return nil 
}
*/