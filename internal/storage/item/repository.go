package item

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "github.com/chrisabs/storage/internal/storage/models"
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
        INSERT INTO item (
            name, description, quantity, container_id, 
            family_id, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at`

    err = tx.QueryRow(
        itemQuery,
        item.Name,
        item.Description,
        item.Quantity,
        item.ContainerID,
        item.FamilyID,
        time.Now().UTC(),
        time.Now().UTC(),
    ).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

    if err != nil {
        return nil, fmt.Errorf("error creating item: %v", err)
    }

    for _, tagName := range tagNames {
        var tagID int
        err := tx.QueryRow(`
            SELECT id FROM tag 
            WHERE name = $1 AND family_id = $2`,
            tagName, item.FamilyID,
        ).Scan(&tagID)

        if err == sql.ErrNoRows {
            err = tx.QueryRow(`
                INSERT INTO tag (name, family_id, created_at, updated_at)
                VALUES ($1, $2, $3, $4)
                RETURNING id`,
                tagName,
                item.FamilyID,
                time.Now().UTC(),
                time.Now().UTC(),
            ).Scan(&tagID)

            if err != nil {
                return nil, fmt.Errorf("error creating tag: %v", err)
            }
        } else if err != nil {
            return nil, fmt.Errorf("error checking existing tag: %v", err)
        }

        _, err = tx.Exec(
            "INSERT INTO item_tag (item_id, tag_id) VALUES ($1, $2)",
            item.ID, tagID,
        )
        if err != nil {
            return nil, fmt.Errorf("error linking tag to item: %v", err)
        }
    }

    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("error committing transaction: %v", err)
    }

    return r.GetByID(item.ID, item.FamilyID)
}

func (r *Repository) GetByID(id int, familyID int) (*models.Item, error) {
    query := `
        WITH item_images AS (
            SELECT item_id,
                   jsonb_agg(
                       jsonb_build_object(
                           'url', url,
                           'displayOrder', display_order,
                           'createdAt', created_at,
                           'updatedAt', updated_at
                       ) ORDER BY display_order
                   ) as images
            FROM item_image
            GROUP BY item_id
        )
        SELECT i.id, i.name, i.description, i.quantity, 
               i.container_id, i.family_id, i.created_at, i.updated_at,
               COALESCE(img.images, '[]'::jsonb) as images,
               COALESCE(
                    jsonb_build_object(
                        'id', c.id,
                        'name', c.name,
                        'description', c.description,
                        'qrCode', c.qr_code,
                        'qrCodeImage', c.qr_code_image,
                        'number', c.number,
                        'location', c.location,
                        'userId', c.user_id,
                        'familyId', c.family_id,
                        'workspaceId', c.workspace_id,
                        'workspace', CASE 
                            WHEN w.id IS NOT NULL THEN
                                jsonb_build_object(
                                    'id', w.id,
                                    'name', w.name,
                                    'description', w.description,
                                    'userId', w.user_id,
                                    'familyId', w.family_id,
                                    'createdAt', w.created_at,
                                    'updatedAt', w.updated_at
                                )
                            ELSE null
                        END,
                        'createdAt', c.created_at,
                        'updatedAt', c.updated_at
                    ),
                    null
                ) as container,
               COALESCE(
                   jsonb_agg(
                       DISTINCT jsonb_build_object(
                           'id', t.id,
                           'name', t.name,
                           'colour', t.colour,
                           'familyId', t.family_id,
                           'createdAt', t.created_at,
                           'updatedAt', t.updated_at
                       )
                   ) FILTER (WHERE t.id IS NOT NULL),
                   '[]'
               ) as tags
        FROM item i
        LEFT JOIN item_images img ON i.id = img.item_id
        LEFT JOIN container c ON i.container_id = c.id AND c.family_id = i.family_id
        LEFT JOIN workspace w ON c.workspace_id = w.id AND w.family_id = c.family_id
        LEFT JOIN item_tag it ON i.id = it.item_id
        LEFT JOIN tag t ON it.tag_id = t.id AND t.family_id = i.family_id
        WHERE i.id = $1 AND i.family_id = $2
        GROUP BY i.id, i.name, i.description, i.quantity, 
                 i.container_id, i.family_id, i.created_at, i.updated_at,
                 img.images,
                 c.id, c.name, c.description, c.qr_code, c.qr_code_image, c.number, c.location,
                 c.user_id, c.family_id, c.workspace_id, c.created_at, c.updated_at,
                 w.id, w.name, w.description, w.user_id, w.family_id, w.created_at, w.updated_at`

    item := new(models.Item)
    var imagesJSON, containerJSON, tagsJSON []byte

    err := r.db.QueryRow(query, id, familyID).Scan(
        &item.ID, &item.Name, &item.Description,
        &item.Quantity, &item.ContainerID, &item.FamilyID,
        &item.CreatedAt, &item.UpdatedAt,
        &imagesJSON, &containerJSON, &tagsJSON,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("item not found")
    }
    if err != nil {
        return nil, err
    }

    if err := json.Unmarshal(imagesJSON, &item.Images); err != nil {
        return nil, fmt.Errorf("error parsing images: %v", err)
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

func (r *Repository) GetByFamilyID(familyID int) ([]*models.Item, error) {
    query := `
        WITH item_images AS (
            SELECT item_id,
                   jsonb_agg(
                       jsonb_build_object(
                           'url', url,
                           'displayOrder', display_order,
                           'createdAt', created_at,
                           'updatedAt', updated_at
                       ) ORDER BY display_order
                   ) as images
            FROM item_image
            GROUP BY item_id
        )
        SELECT i.id, i.name, i.description, i.quantity, 
               i.container_id, i.family_id, i.created_at, i.updated_at,
               COALESCE(img.images, '[]'::jsonb) as images,
               COALESCE(
                    jsonb_build_object(
                        'id', c.id,
                        'name', c.name,
                        'description', c.description,
                        'qrCode', c.qr_code,
                        'qrCodeImage', c.qr_code_image,
                        'number', c.number,
                        'location', c.location,
                        'userId', c.user_id,
                        'familyId', c.family_id,
                        'workspaceId', c.workspace_id,
                        'workspace', CASE 
                            WHEN w.id IS NOT NULL THEN
                                jsonb_build_object(
                                    'id', w.id,
                                    'name', w.name,
                                    'description', w.description,
                                    'userId', w.user_id,
                                    'familyId', w.family_id,
                                    'createdAt', w.created_at,
                                    'updatedAt', w.updated_at
                                )
                            ELSE null
                        END,
                        'createdAt', c.created_at,
                        'updatedAt', c.updated_at
                    ),
                    null
                ) as container,
               COALESCE(
                   jsonb_agg(
                       DISTINCT jsonb_build_object(
                           'id', t.id,
                           'name', t.name,
                           'colour', t.colour,
                           'familyId', t.family_id,
                           'createdAt', t.created_at,
                           'updatedAt', t.updated_at
                       )
                   ) FILTER (WHERE t.id IS NOT NULL),
                   '[]'
               ) as tags
        FROM item i
        LEFT JOIN item_images img ON i.id = img.item_id
        LEFT JOIN container c ON i.container_id = c.id AND c.family_id = i.family_id
        LEFT JOIN workspace w ON c.workspace_id = w.id AND w.family_id = c.family_id
        LEFT JOIN item_tag it ON i.id = it.item_id
        LEFT JOIN tag t ON it.tag_id = t.id AND t.family_id = i.family_id
        WHERE i.family_id = $1
        GROUP BY i.id, i.name, i.description, i.quantity, 
                 i.container_id, i.family_id, i.created_at, i.updated_at,
                 img.images,
                 c.id, c.name, c.description, c.qr_code, c.qr_code_image, c.number, c.location,
                 c.user_id, c.family_id, c.workspace_id, c.created_at, c.updated_at,
                 w.id, w.name, w.description, w.user_id, w.family_id, w.created_at, w.updated_at
        ORDER BY i.created_at DESC`

    rows, err := r.db.Query(query, familyID)
    if err != nil {
        return nil, fmt.Errorf("error querying items: %v", err)
    }
    defer rows.Close()

    var items []*models.Item
    for rows.Next() {
        item := new(models.Item)
        var imagesJSON, containerJSON, tagsJSON []byte

        err := rows.Scan(
            &item.ID, &item.Name, &item.Description,
            &item.Quantity, &item.ContainerID, &item.FamilyID,
            &item.CreatedAt, &item.UpdatedAt,
            &imagesJSON, &containerJSON, &tagsJSON,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning item: %v", err)
        }

        if err := json.Unmarshal(imagesJSON, &item.Images); err != nil {
            return nil, fmt.Errorf("error parsing images: %v", err)
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
        SET name = $2, description = $3,
            quantity = $4, container_id = $5, updated_at = $6
        WHERE id = $1 AND family_id = $7`

    result, err := tx.Exec(
        query,
        item.ID,
        item.Name,
        item.Description,
        item.Quantity,
        item.ContainerID,
        time.Now().UTC(),
        item.FamilyID,
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

func (r *Repository) AddItemImage(itemID int, familyID int, url string, displayOrder int) error {
    query := `
        INSERT INTO item_image (item_id, url, display_order)
        SELECT $1, $2, $3
        FROM item
        WHERE id = $1 AND family_id = $4`
    
    result, err := r.db.Exec(query, itemID, url, displayOrder, familyID)
    if err != nil {
        return fmt.Errorf("error adding item image: %v", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking result: %v", err)
    }

    if rows == 0 {
        return fmt.Errorf("item not found or access denied")
    }
    
    return nil
}

func (r *Repository) DeleteItemImage(itemID int, familyID int, url string) error {
    query := `
        DELETE FROM item_image
        WHERE item_id = $1 AND url = $2
        AND EXISTS (
            SELECT 1 FROM item
            WHERE id = $1 AND family_id = $3
        )`
    
    result, err := r.db.Exec(query, itemID, url, familyID)
    if err != nil {
        return fmt.Errorf("error deleting item image: %v", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking delete result: %v", err)
    }
    
    if rowsAffected == 0 {
        return fmt.Errorf("image not found or access denied")
    }
    
    return nil
}

func (r *Repository) Delete(id int, familyID int) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    imageQuery := `
        DELETE FROM item_image
        WHERE item_id = $1
        AND EXISTS (
            SELECT 1 FROM item
            WHERE id = $1 AND family_id = $2
        )`
    
    _, err = tx.Exec(imageQuery, id, familyID)
    if err != nil {
        return fmt.Errorf("error removing item images: %v", err)
    }

    itemTagQuery := `
        DELETE FROM item_tag
        WHERE item_id = $1
        AND EXISTS (
            SELECT 1 FROM item
            WHERE id = $1 AND family_id = $2
        )`
    
    _, err = tx.Exec(itemTagQuery, id, familyID)
    if err != nil {
        return fmt.Errorf("error removing item-tag associations: %v", err)
    }

    itemQuery := `DELETE FROM item WHERE id = $1 AND family_id = $2`
    result, err := tx.Exec(itemQuery, id, familyID)
    if err != nil {
        return fmt.Errorf("error deleting item: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking delete result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("item not found or access denied")
    }

    return tx.Commit()
}