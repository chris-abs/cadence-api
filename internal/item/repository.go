package item

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

func (r *Repository) Create(item *models.Item) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	itemQuery := `
        INSERT INTO item (name, description, image_url, quantity, container_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`

	var itemID int
	err = tx.QueryRow(
		itemQuery,
		item.Name,
		item.Description,
		item.ImageURL,
		item.Quantity,
		item.ContainerID,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&itemID)

	if err != nil {
		return fmt.Errorf("error creating item: %v", err)
	}

	return tx.Commit()
}

func (r *Repository) GetByID(id int) (*models.Item, error) {
	query := `
        SELECT id, name, description, image_url, quantity, 
               container_id, created_at, updated_at
        FROM item
        WHERE id = $1`

	item := new(models.Item)
	err := r.db.QueryRow(query, id).Scan(
		&item.ID, &item.Name, &item.Description, &item.ImageURL,
		&item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("item not found")
	}
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (r *Repository) GetByUserID(userID int) ([]*models.Item, error) {
	query := `
        SELECT i.id, i.name, i.description, i.image_url, i.quantity, 
               i.container_id, i.created_at, i.updated_at
        FROM item i
        JOIN container c ON i.container_id = c.id
        WHERE c.user_id = $1
        ORDER BY i.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		item := new(models.Item)
		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.ImageURL,
			&item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) Update(item *models.Item) error {
	query := `
        UPDATE item
        SET name = $2, description = $3, image_url = $4,
            quantity = $5, container_id = $6, updated_at = $7
        WHERE id = $1`

	result, err := r.db.Exec(
		query,
		item.ID,
		item.Name,
		item.Description,
		item.ImageURL,
		item.Quantity,
		item.ContainerID,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("error updating item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found")
	}

	return nil
}

func (r *Repository) Delete(id int) error {
	query := `DELETE FROM item WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found")
	}

	return nil
}
