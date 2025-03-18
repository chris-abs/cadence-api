package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/chrisabs/cadence/internal/platform/database/migrations"
)

type PostgresDB struct {
    *sql.DB
    migrationsManager *migrations.Manager
}

func NewPostgresDB() (*PostgresDB, error) {
    password := os.Getenv("POSTGRES_PASSWORD")

    username := os.Getenv("POSTGRES_USER")
    if username == "" {
        username = "postgres" 
    }
    
    connStr := fmt.Sprintf(
        "host=localhost user=%s dbname=postgres password=%s sslmode=disable port=5432",
        username, password,
    )
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("error connecting to database: %v", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("error pinging database: %v", err)
    }

    postgresDB := &PostgresDB{DB: db}
    postgresDB.migrationsManager = migrations.NewManager(db)

    return postgresDB, nil
}