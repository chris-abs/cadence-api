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