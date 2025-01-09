package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	*sql.DB
}

func NewPostgresDB() (*PostgresDB, error) {
	connStr := "user=postgres dbname=postgres password=STQRAGE sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return &PostgresDB{DB: db}, nil
}

func (db *PostgresDB) Init() error {
	fmt.Println("Starting database initialization...")

	// Comment out the table checking and dropping section
	/*
	   // Check existing tables
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
	   }
	*/

	// Continue with ensuring tables exist
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

	return nil
}

func (db *PostgresDB) createUsersTable() error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        password TEXT NOT NULL,
        first_name VARCHAR(100),
        last_name VARCHAR(100),
        image_url TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
        `
	_, err := db.Exec(query)
	return err
}

func (db *PostgresDB) createWorkspaceTable() error {
    query := `
        CREATE TABLE IF NOT EXISTS workspace (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            description TEXT,
            user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS workspace_container (
            workspace_id INTEGER REFERENCES workspace(id) ON DELETE CASCADE,
            container_id INTEGER REFERENCES container(id) ON DELETE CASCADE,
            PRIMARY KEY (workspace_id, container_id)
        );

        CREATE INDEX IF NOT EXISTS idx_workspace_user ON workspace(user_id);
        CREATE INDEX IF NOT EXISTS idx_workspace_container_workspace ON workspace_container(workspace_id);
        CREATE INDEX IF NOT EXISTS idx_workspace_container_container ON workspace_container(container_id);
    `

    _, err := db.Exec(query)
    if err != nil {
        return fmt.Errorf("error creating workspace tables: %v", err)
    }

    return nil
}

func (db *PostgresDB) createContainerTable() error {
	query := `
        CREATE TABLE IF NOT EXISTS container (
            id SERIAL PRIMARY KEY,
            name VARCHAR(50),
            qr_code VARCHAR(100) UNIQUE,           
            qr_code_image TEXT,             
            number INTEGER,         
            location VARCHAR(50),
            user_id INTEGER,        
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );

        CREATE INDEX IF NOT EXISTS idx_container_qr_code ON container(qr_code);
    `

	// DROP container TABLE FOR TESTING
	// dropQuery := `DROP TABLE IF EXISTS container;`
	// if _, err := db.Exec(dropQuery); err != nil {
	//     return fmt.Errorf("error dropping table: %v", err)
	// }

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error executing create table query: %v", err)
	}

	return nil
}

func (db *PostgresDB) createItemTables() error {
	query := `
        CREATE TABLE IF NOT EXISTS tag (
            id SERIAL PRIMARY KEY,
            name VARCHAR(50) UNIQUE,
            colour TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS item (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100),
            description TEXT,
            image_url TEXT,
            quantity INTEGER,
            container_id INTEGER REFERENCES container(id) ON DELETE CASCADE NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS item_tag (
            item_id INTEGER REFERENCES item(id) ON DELETE CASCADE,
            tag_id INTEGER REFERENCES tag(id) ON DELETE CASCADE,
            PRIMARY KEY (item_id, tag_id)
        );

        CREATE INDEX IF NOT EXISTS idx_item_container ON item(container_id);
        CREATE INDEX IF NOT EXISTS idx_item_tag_item ON item_tag(item_id);
        CREATE INDEX IF NOT EXISTS idx_item_tag_tag ON item_tag(tag_id);
    `
	// DROP item TABLE FOR TESTING
	// dropQuery := `
	//     DROP TABLE IF EXISTS item_tag;
	//     DROP TABLE IF EXISTS item;
	//     DROP TABLE IF EXISTS tag;
	// `
	// if _, err := db.Exec(dropQuery); err != nil {
	// 	return fmt.Errorf("error dropping table: %v", err)
	// }

	_, err := db.Exec(query)
	return err
}
