package container

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/storage/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(container *models.Container) error {
	query := `
        INSERT INTO container (id, name, qr_code, qr_code_image, number, location, user_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id`

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

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
		return fmt.Errorf("error creating container: %v", err)
	}

	return tx.Commit()
}

func (r *Repository) GetByID(id int) (*models.Container, error) {
	query := `
        SELECT c.id, c.name, c.qr_code, c.qr_code_image, c.number, 
               c.location, c.user_id, c.created_at, c.updated_at
        FROM container c
        WHERE c.id = $1`

	container := new(models.Container)
	err := r.db.QueryRow(query, id).Scan(
		&container.ID, &container.Name, &container.QRCode,
		&container.QRCodeImage, &container.Number, &container.Location,
		&container.UserID, &container.CreatedAt, &container.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("container not found")
	}
	if err != nil {
		return nil, err
	}

	itemsQuery := `
        SELECT id, name, description, image_url, quantity, 
               container_id, created_at, updated_at
        FROM item 
        WHERE container_id = $1`

	rows, err := r.db.Query(itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.ImageURL,
			&item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	container.Items = items
	return container, nil
}

func (r *Repository) GetByUserID(userID int) ([]*models.Container, error) {
	query := `
        SELECT id, name, qr_code, qr_code_image, number, location, 
               user_id, created_at, updated_at 
        FROM container
        WHERE user_id = $1
        ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying containers: %v", err)
	}
	defer rows.Close()

	var containers []*models.Container
	for rows.Next() {
		container := new(models.Container)
		err := rows.Scan(
			&container.ID, &container.Name, &container.QRCode,
			&container.QRCodeImage, &container.Number, &container.Location,
			&container.UserID, &container.CreatedAt, &container.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning container: %v", err)
		}
		containers = append(containers, container)
	}

	return containers, nil
}

func (r *Repository) GetByQR(qrCode string) (*models.Container, error) {
	query := `
        SELECT id, name, qr_code, qr_code_image, number, location, 
               user_id, created_at, updated_at 
        FROM container
        WHERE qr_code = $1`

	container := new(models.Container)
	err := r.db.QueryRow(query, qrCode).Scan(
		&container.ID, &container.Name, &container.QRCode,
		&container.QRCodeImage, &container.Number, &container.Location,
		&container.UserID, &container.CreatedAt, &container.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("container not found")
	}
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (r *Repository) Update(container *models.Container) error {
	query := `
        UPDATE container
        SET name = $2, location = $3, updated_at = $4
        WHERE id = $1`

	result, err := r.db.Exec(
		query,
		container.ID,
		container.Name,
		container.Location,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("error updating container: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("container not found")
	}

	return nil
}

func (r *Repository) Delete(id int) error {
	query := `DELETE FROM container WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting container: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("container not found")
	}

	return nil
}
