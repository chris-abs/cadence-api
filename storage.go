package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateContainer(*Container) error
	DeleteContainer(int) error
	UpdateContainer(*Container) error
	GetContainers() ([]*Container, error)
	GetContainerByID(int) (*Container, error)
}

type PostgresStore struct {
	db *sql.DB
}

func (s *PostgresStore) CreateContainer(container *Container) error {
	query := `
        INSERT INTO container (id, name, qr_code, number, location, user_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id`

	resp, err := s.db.Query(
		query,
		container.ID,
		container.Name,
		container.QRCode,
		container.Number,
		container.Location,
		container.UserId,
		container.CreatedAt,
		container.UpdatedAt,
	)

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)
	return nil
}

func (s *PostgresStore) UpdateContainer(container *Container) error {
	return nil
}

func (s *PostgresStore) DeleteContainer(id int) error {
	return nil
}

func (s *PostgresStore) GetContainerByID(id int) (*Container, error) {
	rows, err := s.db.Query(`
        SELECT id, name, qr_code, number, location, user_id, created_at, updated_at 
        FROM container 
        WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		return scanIntoContainer(rows)
	}

	return nil, fmt.Errorf("container %d not found", id)
}

func (s *PostgresStore) GetContainers() ([]*Container, error) {
	rows, err := s.db.Query(`
        SELECT id, name, qr_code, number, location, user_id, created_at, updated_at 
        FROM container
    `)
	if err != nil {
		return nil, fmt.Errorf("error querying containers: %v", err)
	}
	defer rows.Close()

	containers := []*Container{}
	for rows.Next() {
		container, err := scanIntoContainer(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning container row: %v", err)
		}
		containers = append(containers, container)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return containers, nil
}

func scanIntoContainer(rows *sql.Rows) (*Container, error) {
	container := new(Container)
	err := rows.Scan(
		&container.ID,
		&container.Name,
		&container.QRCode,
		&container.Number,
		&container.Location,
		&container.UserId,
		&container.CreatedAt,
		&container.UpdatedAt,
	)
	return container, err
}

func NewPostgressStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=STQRAGE sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createContainerTable()
}

func (s *PostgresStore) createContainerTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS container (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50),
			qr_code VARCHAR(50),
			number INTEGER,         
			location VARCHAR(50),
			user_id INTEGER,        
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := s.db.Exec(query)
	return err
}
