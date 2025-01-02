package container

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(container *Container, itemIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := `
        INSERT INTO container (id, name, qr_code, qr_code_image, number, location, user_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id`

	var containerID int
	err = tx.QueryRow(
		query,
		container.ID,
		container.Name,
		container.QRCode,
		container.QRCodeImage,
		container.Number,
		container.Location,
		container.UserID,
		container.CreatedAt,
		container.UpdatedAt,
	).Scan(&containerID)

	if err != nil {
		tx.Rollback()
		return err
	}

	if len(itemIDs) > 0 {
		_, err = tx.Exec("UPDATE item SET container_id = $1 WHERE id = ANY($2)", containerID, pq.Array(itemIDs))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) Update(container *Container, itemIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := `
        UPDATE container
        SET name = $2, location = $3, updated_at = $4
        WHERE id = $1
    `
	_, err = tx.Exec(query, container.ID, container.Name, container.Location, time.Now().UTC())
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("UPDATE item SET container_id = NULL WHERE container_id = $1", container.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if len(itemIDs) > 0 {
		_, err = tx.Exec("UPDATE item SET container_id = $1 WHERE id = ANY($2)", container.ID, pq.Array(itemIDs))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) UpdateContainerItems(containerID int, itemIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE item SET container_id = NULL WHERE container_id = $1", containerID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, itemID := range itemIDs {
		_, err = tx.Exec("UPDATE item SET container_id = $1 WHERE id = $2", containerID, itemID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) Delete(id int) error {
	_, err := r.db.Query("DELETE FROM container WHERE id = $1", id)
	return err
}

func (r *Repository) GetByID(id int) (*Container, error) {
	stmt, err := r.db.Prepare(`
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

func (r *Repository) GetByQR(qrCode string) (*Container, error) {
	stmt, err := r.db.Prepare(`
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

func (r *Repository) GetAll() ([]*Container, error) {
	stmt, err := r.db.Prepare(`
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
		&container.UserID,
		&container.CreatedAt,
		&container.UpdatedAt,
	)
	return container, err
}
