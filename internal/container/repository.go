package container

import (
	"database/sql"
	"encoding/json"
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

func (r *Repository) Create(container *models.Container, itemRequests []CreateItemRequest) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    containerQuery := `
        INSERT INTO container (id, name, qr_code, qr_code_image, number, location, user_id, workspace_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id`

    var containerID int
    err = tx.QueryRow(
        containerQuery,
        container.ID,
        container.Name,
        container.QRCode,
        container.QRCodeImage,
        container.Number,
        container.Location,
        container.UserID,
        container.WorkspaceID,
        container.CreatedAt,
        container.UpdatedAt,
    ).Scan(&containerID)

    if err != nil {
        return fmt.Errorf("error creating container: %v", err)
    }

	if len(itemRequests) > 0 {
		itemQuery := `
            INSERT INTO item (name, description, image_url, quantity, container_id, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING id`

		for _, itemReq := range itemRequests {
			var itemID int
			err = tx.QueryRow(
				itemQuery,
				itemReq.Name,
				itemReq.Description,
				itemReq.ImageURL,
				itemReq.Quantity,
				containerID,
				time.Now().UTC(),
				time.Now().UTC(),
			).Scan(&itemID)

			if err != nil {
				return fmt.Errorf("error creating item: %v", err)
			}
		}
	}

	return tx.Commit()
}

func (r *Repository) GetByID(id int) (*models.Container, error) {
    containerQuery := `
        SELECT c.id, c.name, c.qr_code, c.qr_code_image, c.number, 
               c.location, c.user_id, c.workspace_id, c.created_at, c.updated_at
        FROM container c
        WHERE c.id = $1`

    container := new(models.Container)
    err := r.db.QueryRow(containerQuery, id).Scan(
        &container.ID, &container.Name, &container.QRCode,
        &container.QRCodeImage, &container.Number, &container.Location,
        &container.UserID, &container.WorkspaceID, &container.CreatedAt, &container.UpdatedAt,
    )

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("container not found")
	}
	if err != nil {
		return nil, err
	}

	itemsQuery := `
        SELECT i.id, i.name, i.description, i.image_url, i.quantity, 
               i.container_id, i.created_at, i.updated_at,
               COALESCE(
                   jsonb_agg(
                       jsonb_build_object(
                           'id', t.id,
                           'name', t.name,
                           'colour', t.colour,
                           'createdAt', t.created_at AT TIME ZONE 'UTC',
                           'updatedAt', t.updated_at AT TIME ZONE 'UTC'
                       )
                   ) FILTER (WHERE t.id IS NOT NULL),
                   '[]'
               ) as tags
        FROM item i
        LEFT JOIN item_tag it ON i.id = it.item_id
        LEFT JOIN tag t ON it.tag_id = t.id
        WHERE i.container_id = $1
        GROUP BY i.id`

	rows, err := r.db.Query(itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	container.Items = make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		var tagsJSON []byte

		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.ImageURL,
			&item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
			&tagsJSON,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(tagsJSON, &item.Tags); err != nil {
			return nil, fmt.Errorf("error parsing tags: %v", err)
		}

		container.Items = append(container.Items, item)
	}

	return container, nil
}

func (r *Repository) GetByUserID(userID int) ([]*models.Container, error) {
    query := `
        SELECT id, name, qr_code, qr_code_image, number, location, 
               user_id, workspace_id, created_at, updated_at 
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
            &container.UserID, &container.WorkspaceID, &container.CreatedAt, &container.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning container: %v", err)
        }

		itemsQuery := `
            SELECT i.id, i.name, i.description, i.image_url, i.quantity, 
                   i.container_id, i.created_at, i.updated_at,
                   COALESCE(
                       jsonb_agg(
                           jsonb_build_object(
                               'id', t.id,
                               'name', t.name,
                               'colour', t.colour,
                               'createdAt', t.created_at AT TIME ZONE 'UTC',
                               'updatedAt', t.updated_at AT TIME ZONE 'UTC'
                           )
                       ) FILTER (WHERE t.id IS NOT NULL),
                       '[]'
                   ) as tags
            FROM item i
            LEFT JOIN item_tag it ON i.id = it.item_id
            LEFT JOIN tag t ON it.tag_id = t.id
            WHERE i.container_id = $1
            GROUP BY i.id`

		itemRows, err := r.db.Query(itemsQuery, container.ID)
		if err != nil {
			return nil, fmt.Errorf("error querying items: %v", err)
		}

		container.Items = make([]models.Item, 0)
		func() {
			defer itemRows.Close()
			for itemRows.Next() {
				var item models.Item
				var tagsJSON []byte

				err := itemRows.Scan(
					&item.ID, &item.Name, &item.Description, &item.ImageURL,
					&item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
					&tagsJSON,
				)
				if err != nil {
					return
				}

				if err := json.Unmarshal(tagsJSON, &item.Tags); err != nil {
					return
				}

				container.Items = append(container.Items, item)
			}
		}()

		containers = append(containers, container)
	}

	return containers, nil
}

func (r *Repository) GetByQR(qrCode string) (*models.Container, error) {
    query := `
        SELECT id, name, qr_code, qr_code_image, number, location, 
               user_id, workspace_id, created_at, updated_at 
        FROM container
        WHERE qr_code = $1`

    container := new(models.Container)
    err := r.db.QueryRow(query, qrCode).Scan(
        &container.ID, &container.Name, &container.QRCode,
        &container.QRCodeImage, &container.Number, &container.Location,
        &container.UserID, &container.WorkspaceID, &container.CreatedAt, &container.UpdatedAt,
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

	rows, err := r.db.Query(itemsQuery, container.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	container.Items = make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.ImageURL,
			&item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		container.Items = append(container.Items, item)
	}

	return container, nil
}

func (r *Repository) Update(container *models.Container) error {
    query := `
        UPDATE container
        SET name = $2, location = $3, workspace_id = $4, updated_at = $5
        WHERE id = $1`

    var workspaceID sql.NullInt64
    if container.WorkspaceID != nil {
        workspaceID = sql.NullInt64{Int64: int64(*container.WorkspaceID), Valid: true}
    } else {
        workspaceID = sql.NullInt64{Valid: false} 
    }

    result, err := r.db.Exec(
        query,
        container.ID,
        container.Name,
        container.Location,
        workspaceID,
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
