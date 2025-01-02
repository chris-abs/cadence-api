package container

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/storage/internal/item"
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
	query := `
        SELECT c.id, c.name, c.qr_code, c.qr_code_image, c.number, 
               c.location, c.user_id, c.created_at, c.updated_at
        FROM container c
        WHERE c.id = $1`

	container := new(Container)
	err := r.db.QueryRow(query, id).Scan(
		&container.ID, &container.Name, &container.QRCode,
		&container.QRCodeImage, &container.Number, &container.Location,
		&container.UserID, &container.CreatedAt, &container.UpdatedAt,
	)

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

	var items []item.Item
	for rows.Next() {
		var item item.Item
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
	query := `
        SELECT id, name, qr_code, qr_code_image, number, 
               location, user_id, created_at, updated_at 
        FROM container
        ORDER BY created_at DESC`

	containers, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying containers: %v", err)
	}
	defer containers.Close()

	var results []*Container
	for containers.Next() {
		container := new(Container)
		err := containers.Scan(
			&container.ID, &container.Name, &container.QRCode,
			&container.QRCodeImage, &container.Number, &container.Location,
			&container.UserID, &container.CreatedAt, &container.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning container: %v", err)
		}

		itemsQuery := `
            SELECT id, name, description, image_url, quantity, 
                   container_id, created_at, updated_at
            FROM item 
            WHERE container_id = $1`

		items, err := r.db.Query(itemsQuery, container.ID)
		if err != nil {
			return nil, fmt.Errorf("error querying items for container %d: %v", container.ID, err)
		}
		defer items.Close()

		var containerItems []item.Item
		for items.Next() {
			var item item.Item
			err := items.Scan(
				&item.ID, &item.Name, &item.Description, &item.ImageURL,
				&item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("error scanning item: %v", err)
			}
			containerItems = append(containerItems, item)
		}

		container.Items = containerItems
		results = append(results, container)
	}

	if err = containers.Err(); err != nil {
		return nil, fmt.Errorf("error iterating containers: %v", err)
	}

	return results, nil
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
