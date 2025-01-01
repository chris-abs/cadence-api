package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateContainer(*Container) error
	DeleteContainer(int) error
	UpdateContainer(*Container) error
	GetContainers() ([]*Container, error)
	GetContainerByID(int) (*Container, error)
	GetContainerByQR(string) (*Container, error)
}

type PostgresStore struct {
	db *sql.DB
}

func (s *PostgresStore) CreateContainer(container *Container) error {
	query := `
        INSERT INTO container (id, name, qr_code, qr_code_image, number, location, user_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id`

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	resp, err := tx.Query(
		query,
		container.ID,
		container.Name,
		container.QRCode,
		container.QRCodeImage,
		container.Number,
		container.Location,
		container.UserId,
		container.CreatedAt,
		container.UpdatedAt,
	)

	if err != nil {
		tx.Rollback()
		return err
	}
	defer resp.Close()

	return tx.Commit()
}

func (s *PostgresStore) UpdateContainer(container *Container) error {
	query := `
        UPDATE container
        SET name = $2, location = $3, qr_code_image = $4, updated_at = $5
        WHERE id = $1
    `
	_, err := s.db.Exec(
		query,
		container.ID,
		container.Name,
		container.Location,
		container.QRCodeImage,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("error updating container: %v", err)
	}

	return nil
}

func (s *PostgresStore) DeleteContainer(id int) error {
	_, err := s.db.Query("DELETE FROM container WHERE ID = $1", id)
	return err
}

func (s *PostgresStore) GetContainerByID(id int) (*Container, error) {
	stmt, err := s.db.Prepare(`
        SELECT id, name, qr_code, qr_code_image, number, location, user_id, created_at, updated_at 
        FROM container 
        WHERE id = $1
    `)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		return scanIntoContainer(rows)
	}

	return nil, fmt.Errorf("container %d not found", id)
}

func (s *PostgresStore) GetContainerByQR(qrCode string) (*Container, error) {
	stmt, err := s.db.Prepare(`
        SELECT id, name, qr_code, qr_code_image, number, location, user_id, created_at, updated_at 
        FROM container 
        WHERE qr_code = $1
    `)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(qrCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		return scanIntoContainer(rows)
	}

	return nil, fmt.Errorf("container with QR code %s not found", qrCode)
}

func (s *PostgresStore) GetContainers() ([]*Container, error) {
	stmt, err := s.db.Prepare(`
        SELECT id, name, qr_code, qr_code_image, number, location, user_id, created_at, updated_at 
        FROM container
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
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
		&container.QRCodeImage,
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
	// DROPS TABLE WHEN CREATING A NEW CONTAINER
	// _, err := s.db.Exec(`DROP TABLE IF EXISTS container;`)
	// if err != nil {
	// 	return fmt.Errorf("error dropping table: %v", err)
	// }
	return s.createContainerTable()
}

func (s *PostgresStore) createContainerTable() error {
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
	_, err := s.db.Exec(query)
	return err
}
