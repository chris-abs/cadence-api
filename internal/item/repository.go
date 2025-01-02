package item

import (
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(item *Item, tagIDs []int) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var itemID int
	query := `
        INSERT INTO item (name, description, image_url, quantity, container_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`

	err = tx.QueryRow(
		query,
		item.Name,
		item.Description,
		item.ImageURL,
		item.Quantity,
		item.ContainerID,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&itemID)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			_, err = tx.Exec("INSERT INTO item_tag (item_id, tag_id) VALUES ($1, $2)", itemID, tagID)
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return itemID, nil
}

func (r *Repository) GetByID(id int) (*Item, error) {
	query := `
        SELECT id, name, description, image_url, quantity, 
               container_id, created_at, updated_at
        FROM item 
        WHERE id = $1`

	item := new(Item)
	err := r.db.QueryRow(query, id).Scan(
		&item.ID, &item.Name, &item.Description, &item.ImageURL,
		&item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("item not found")
	}
	return item, err
}

// func (r *Repository) GetItemTags(itemID int) ([]Tag, error) {
//     query := `
//         SELECT t.id, t.name, t.created_at, t.updated_at
//         FROM tag t
//         JOIN item_tag it ON t.id = it.tag_id
//         WHERE it.item_id = $1`

// }

func (r *Repository) GetByUserID(userID int) ([]*Item, error) {
	query := `
        SELECT i.id, i.name, i.description, i.image_url, i.quantity, 
               i.container_id, i.created_at, i.updated_at
        FROM item i
        JOIN container c ON i.container_id = c.id
        WHERE c.user_id = $1
        ORDER BY i.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying items: %v", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item := new(Item)
		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.ImageURL,
			&item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning item: %v", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) GetAll() ([]*Item, error) {
	query := `
        SELECT i.id, i.name, i.description, i.image_url, i.quantity, i.container_id, i.created_at, i.updated_at
        FROM item i
        ORDER BY i.created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item := &Item{}
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

func (r *Repository) Update(item *Item, tagIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := `
        UPDATE item
        SET name = $2, description = $3, image_url = $4, quantity = $5,
            container_id = $6, updated_at = $7
        WHERE id = $1`

	_, err = tx.Exec(
		query,
		item.ID, item.Name, item.Description, item.ImageURL,
		item.Quantity, item.ContainerID, time.Now().UTC(),
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	// Update tags
	_, err = tx.Exec("DELETE FROM item_tag WHERE item_id = $1", item.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, tagID := range tagIDs {
		_, err = tx.Exec("INSERT INTO item_tag (item_id, tag_id) VALUES ($1, $2)", item.ID, tagID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM item WHERE id = $1", id)
	return err
}
