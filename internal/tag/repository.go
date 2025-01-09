package tag

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

func (r *Repository) Create(tag *models.Tag) error {
	query := `
        INSERT INTO tag (name, colour, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id`

	err := r.db.QueryRow(
		query,
		tag.Name,
		tag.Colour,
		tag.CreatedAt,
		tag.UpdatedAt,
	).Scan(&tag.ID)

	if err != nil {
		return fmt.Errorf("error creating tag: %v", err)
	}

	return nil
}

func (r *Repository) GetByID(id int) (*models.Tag, error) {
	query := `
        SELECT t.id, t.name, t.colour, t.created_at, t.updated_at
        FROM tag t
        WHERE t.id = $1`

	tag := new(models.Tag)
	err := r.db.QueryRow(query, id).Scan(
		&tag.ID, &tag.Name, &tag.Colour,
		&tag.CreatedAt, &tag.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag not found")
	}
	if err != nil {
		return nil, err
	}

	itemsQuery := `
        SELECT i.id, i.name, i.description, i.image_url, i.quantity, 
               i.container_id, i.created_at, i.updated_at
        FROM item i
        JOIN item_tag it ON i.id = it.item_id
        WHERE it.tag_id = $1`

	rows, err := r.db.Query(itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tag.Items = make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.ImageURL,
			&item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tag.Items = append(tag.Items, item)
	}

	return tag, nil
}

func (r *Repository) GetAll() ([]*models.Tag, error) {
	query := `
        SELECT id, name, COALESCE(colour, '') as colour, created_at, updated_at
        FROM tag
        ORDER BY name ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		tag := new(models.Tag)
		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Colour,
			&tag.CreatedAt, &tag.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		tag.Items = make([]models.Item, 0)
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *Repository) Update(tag *models.Tag) error {
	query := `
        UPDATE tag
        SET name = $2, colour = $3, updated_at = $4
        WHERE id = $1`

	result, err := r.db.Exec(
		query,
		tag.ID,
		tag.Name,
		tag.Colour,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("error updating tag: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}

func (r *Repository) Delete(id int) error {
	query := `DELETE FROM tag WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting tag: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}
