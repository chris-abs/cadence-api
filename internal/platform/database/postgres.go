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
	if err := db.createContainerTable(); err != nil {
		return fmt.Errorf("error creating container table: %v", err)
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
