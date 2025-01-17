package item

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

func (r *Repository) Create(item *models.Item, tagNames []string) (*models.Item, error) {
    tx, err := r.db.Begin()
    if err != nil {
        return nil, fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    itemQuery := `
        INSERT INTO item (name, description, image_url, quantity, container_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at`

    err = tx.QueryRow(
        itemQuery,
        item.Name,
        item.Description,
        item.ImageURL,
        item.Quantity,
        item.ContainerID,
        time.Now().UTC(),
        time.Now().UTC(),
    ).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

    if err != nil {
        return nil, fmt.Errorf("error creating item: %v", err)
    }

    for _, tagName := range tagNames {
        var tagID int
        err := tx.QueryRow("SELECT id FROM tag WHERE name = $1", tagName).Scan(&tagID)

        if err == sql.ErrNoRows {
            err = tx.QueryRow(`
                INSERT INTO tag (name, created_at, updated_at)
                VALUES ($1, $2, $3)
                RETURNING id`,
                tagName,
                time.Now().UTC(),
                time.Now().UTC(),
            ).Scan(&tagID)

            if err != nil {
                return nil, fmt.Errorf("error creating tag: %v", err)
            }
        } else if err != nil {
            return nil, fmt.Errorf("error checking existing tag: %v", err)
        }

        _, err = tx.Exec("INSERT INTO item_tag (item_id, tag_id) VALUES ($1, $2)", item.ID, tagID)
        if err != nil {
            return nil, fmt.Errorf("error linking tag to item: %v", err)
        }
    }

    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("error committing transaction: %v", err)
    }

    return r.GetByID(item.ID)
}

func (r *Repository) GetByID(id int) (*models.Item, error) {
    query := `
        SELECT i.id, i.name, i.description, i.image_url, i.quantity, 
               i.container_id, i.created_at, i.updated_at,
               COALESCE(
                    jsonb_build_object(
                        'id', c.id,
                        'name', c.name,
                        'qrCode', c.qr_code,
                        'qrCodeImage', c.qr_code_image,
                        'number', c.number,
                        'location', c.location,
                        'userId', c.user_id,
                        'workspaceId', c.workspace_id,
                        'createdAt', c.created_at AT TIME ZONE 'UTC',
                        'updatedAt', c.updated_at AT TIME ZONE 'UTC'
                    ),
                    null
                ) as container,
               COALESCE(
                   jsonb_agg(
                       DISTINCT jsonb_build_object(
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
        LEFT JOIN container c ON i.container_id = c.id
        LEFT JOIN item_tag it ON i.id = it.item_id
        LEFT JOIN tag t ON it.tag_id = t.id
        WHERE i.id = $1
        GROUP BY i.id, i.name, i.description, i.image_url, i.quantity, 
                 i.container_id, i.created_at, i.updated_at,
                 c.id, c.name, c.location, c.created_at, c.updated_at`

    item := new(models.Item)
    var containerJSON, tagsJSON []byte

    err := r.db.QueryRow(query, id).Scan(
        &item.ID, &item.Name, &item.Description, &item.ImageURL,
        &item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
        &containerJSON, &tagsJSON,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("item not found")
    }
    if err != nil {
        return nil, err
    }

    if containerJSON != nil {
        if err := json.Unmarshal(containerJSON, &item.Container); err != nil {
            return nil, fmt.Errorf("error parsing container: %v", err)
        }
    }

    if err := json.Unmarshal(tagsJSON, &item.Tags); err != nil {
        return nil, fmt.Errorf("error parsing tags: %v", err)
    }

    return item, nil
}

func (r *Repository) GetByUserID(userID int) ([]*models.Item, error) {
    query := `
        SELECT DISTINCT i.id, i.name, i.description, i.image_url, i.quantity, 
               i.container_id, i.created_at, i.updated_at,
               COALESCE(
                    jsonb_build_object(
                        'id', c.id,
                        'name', c.name,
                        'qrCode', c.qr_code,
                        'qrCodeImage', c.qr_code_image,
                        'number', c.number,
                        'location', c.location,
                        'userId', c.user_id,
                        'workspaceId', c.workspace_id,
                        'createdAt', c.created_at AT TIME ZONE 'UTC',
                        'updatedAt', c.updated_at AT TIME ZONE 'UTC'
                    ),
                    null
                ) as container,
               COALESCE(
                   jsonb_agg(
                       DISTINCT jsonb_build_object(
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
        LEFT JOIN container c ON i.container_id = c.id
        LEFT JOIN item_tag it ON i.id = it.item_id
        LEFT JOIN tag t ON it.tag_id = t.id
        WHERE c.user_id = $1 OR c.user_id IS NULL
        GROUP BY i.id, i.name, i.description, i.image_url, i.quantity, 
                 i.container_id, i.created_at, i.updated_at,
                 c.id, c.name, c.location, c.created_at, c.updated_at
        ORDER BY i.created_at DESC`

    rows, err := r.db.Query(query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []*models.Item
    for rows.Next() {
        item := new(models.Item)
        var containerJSON, tagsJSON []byte

        err := rows.Scan(
            &item.ID, &item.Name, &item.Description, &item.ImageURL,
            &item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
            &containerJSON, &tagsJSON,
        )
        if err != nil {
            return nil, err
        }

        if containerJSON != nil {
            if err := json.Unmarshal(containerJSON, &item.Container); err != nil {
                return nil, fmt.Errorf("error parsing container: %v", err)
            }
        }

        if err := json.Unmarshal(tagsJSON, &item.Tags); err != nil {
            return nil, fmt.Errorf("error parsing tags: %v", err)
        }

        items = append(items, item)
    }

    return items, nil
}

func (r *Repository) Update(item *models.Item) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    query := `
        UPDATE item
        SET name = $2, description = $3, image_url = $4,
            quantity = $5, container_id = $6, updated_at = $7
        WHERE id = $1`

    result, err := tx.Exec(
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

    _, err = tx.Exec("DELETE FROM item_tag WHERE item_id = $1", item.ID)
    if err != nil {
        return fmt.Errorf("error removing old tags: %v", err)
    }

    if len(item.Tags) > 0 {
        tagQuery := `INSERT INTO item_tag (item_id, tag_id) VALUES ($1, $2)`
        for _, tag := range item.Tags {
            _, err = tx.Exec(tagQuery, item.ID, tag.ID)
            if err != nil {
                return fmt.Errorf("error associating tag: %v", err)
            }
        }
    }

    return tx.Commit()
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