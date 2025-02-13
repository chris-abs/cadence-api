package tag

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/chrisabs/storage/internal/models"
	"github.com/lib/pq"
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
        WITH item_images AS (
            SELECT item_id,
                   jsonb_agg(
                       jsonb_build_object(
                           'id', id,
                           'url', url,
                           'displayOrder', display_order,
                           'createdAt', created_at,
                           'updatedAt', updated_at
                       ) ORDER BY display_order
                   ) FILTER (WHERE id IS NOT NULL) as images
            FROM item_image
            GROUP BY item_id
        )
        SELECT t.id, t.name, t.colour, t.created_at, t.updated_at,
               COALESCE(
                   jsonb_agg(
                       DISTINCT jsonb_build_object(
                           'id', i.id,
                           'name', i.name,
                           'description', i.description,
                           'images', COALESCE(img.images, '[]'::jsonb),
                           'quantity', i.quantity,
                           'containerId', i.container_id,
                           'container', CASE 
                               WHEN c.id IS NOT NULL THEN
                                   jsonb_build_object(
                                       'id', c.id,
                                       'name', c.name,
                                       'qrCode', c.qr_code,
                                       'qrCodeImage', c.qr_code_image,
                                       'number', c.number,
                                       'location', c.location,
                                       'userId', c.user_id,
                                       'workspaceId', c.workspace_id,
                                       'workspace', CASE 
                                           WHEN w.id IS NOT NULL THEN
                                               jsonb_build_object(
                                                   'id', w.id,
                                                   'name', w.name,
                                                   'description', w.description,
                                                   'userId', w.user_id,
                                                   'createdAt', w.created_at,
                                                   'updatedAt', w.updated_at
                                               )
                                           ELSE null
                                       END,
                                       'createdAt', c.created_at,
                                       'updatedAt', c.updated_at
                                   )
                               ELSE null
                           END,
                           'createdAt', i.created_at,
                           'updatedAt', i.updated_at,
                           'tags', COALESCE(
                               (
                                   SELECT jsonb_agg(
                                       jsonb_build_object(
                                           'id', it_tags.id,
                                           'name', it_tags.name,
                                           'colour', it_tags.colour,
                                           'createdAt', it_tags.created_at,
                                           'updatedAt', it_tags.updated_at
                                       )
                                   )
                                   FROM tag it_tags
                                   JOIN item_tag iit ON iit.tag_id = it_tags.id
                                   WHERE iit.item_id = i.id
                               ),
                               '[]'::jsonb
                           )
                       )
                   ) FILTER (WHERE i.id IS NOT NULL),
                   '[]'
               ) as items
        FROM tag t
        LEFT JOIN item_tag it ON t.id = it.tag_id
        LEFT JOIN item i ON it.item_id = i.id
        LEFT JOIN item_images img ON i.id = img.item_id
        LEFT JOIN container c ON i.container_id = c.id
        LEFT JOIN workspace w ON c.workspace_id = w.id
        WHERE t.id = $1
        GROUP BY t.id, t.name, t.colour, t.created_at, t.updated_at`

    tag := new(models.Tag)
    var itemsJSON []byte

    err := r.db.QueryRow(query, id).Scan(
        &tag.ID, &tag.Name, &tag.Colour,
        &tag.CreatedAt, &tag.UpdatedAt,
        &itemsJSON,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("tag not found")
    }
    if err != nil {
        return nil, err
    }

    if err := json.Unmarshal(itemsJSON, &tag.Items); err != nil {
        return nil, fmt.Errorf("error parsing items: %v", err)
    }

    return tag, nil
}

func (r *Repository) GetAll() ([]*models.Tag, error) {
    query := `
        WITH item_images AS (
            SELECT item_id,
                   jsonb_agg(
                       jsonb_build_object(
                           'id', id,
                           'url', url,
                           'displayOrder', display_order,
                           'createdAt', created_at,
                           'updatedAt', updated_at
                       ) ORDER BY display_order
                   ) as images
            FROM item_image
            GROUP BY item_id
        )
        SELECT t.id, t.name, COALESCE(t.colour, '') as colour, 
               t.created_at, t.updated_at,
               COALESCE(
                   jsonb_agg(
                       DISTINCT jsonb_build_object(
                           'id', i.id,
                           'name', i.name,
                           'description', i.description,
                           'images', COALESCE(img.images, '[]'::jsonb),
                           'quantity', i.quantity,
                           'containerId', i.container_id,
                           'container', CASE 
                               WHEN c.id IS NOT NULL THEN
                                   jsonb_build_object(
                                       'id', c.id,
                                       'name', c.name,
                                       'qrCode', c.qr_code,
                                       'qrCodeImage', c.qr_code_image,
                                       'number', c.number,
                                       'location', c.location,
                                       'userId', c.user_id,
                                       'workspaceId', c.workspace_id,
                                       'workspace', CASE 
                                           WHEN w.id IS NOT NULL THEN
                                               jsonb_build_object(
                                                   'id', w.id,
                                                   'name', w.name,
                                                   'description', w.description,
                                                   'userId', w.user_id,
                                                   'createdAt', w.created_at,
                                                   'updatedAt', w.updated_at
                                               )
                                           ELSE null
                                       END,
                                       'createdAt', c.created_at,
                                       'updatedAt', c.updated_at
                                   )
                               ELSE null
                           END,
                           'createdAt', i.created_at,
                           'updatedAt', i.updated_at,
                           'tags', COALESCE(
                               (
                                   SELECT jsonb_agg(
                                       jsonb_build_object(
                                           'id', it_tags.id,
                                           'name', it_tags.name,
                                           'colour', it_tags.colour,
                                           'createdAt', it_tags.created_at,
                                           'updatedAt', it_tags.updated_at
                                       )
                                   )
                                   FROM tag it_tags
                                   JOIN item_tag iit ON iit.tag_id = it_tags.id
                                   WHERE iit.item_id = i.id
                               ),
                               '[]'::jsonb
                           )
                       )
                   ) FILTER (WHERE i.id IS NOT NULL),
                   '[]'
               ) as items
        FROM tag t
        LEFT JOIN item_tag it ON t.id = it.tag_id
        LEFT JOIN item i ON it.item_id = i.id
        LEFT JOIN item_images img ON i.id = img.item_id
        LEFT JOIN container c ON i.container_id = c.id
        LEFT JOIN workspace w ON c.workspace_id = w.id
        GROUP BY t.id, t.name, t.colour, t.created_at, t.updated_at
        ORDER BY t.name ASC`

    rows, err := r.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tags []*models.Tag
    for rows.Next() {
        tag := new(models.Tag)
        var itemsJSON []byte

        err := rows.Scan(
            &tag.ID, &tag.Name, &tag.Colour,
            &tag.CreatedAt, &tag.UpdatedAt,
            &itemsJSON,
        )
        if err != nil {
            return nil, err
        }

        if err := json.Unmarshal(itemsJSON, &tag.Items); err != nil {
            return nil, fmt.Errorf("error parsing items: %v", err)
        }

        tags = append(tags, tag)
    }

    return tags, nil
}

func (r *Repository) Update(tag *models.Tag) error {
    query := `
        UPDATE tag
        SET name = $2, 
            colour = $3, 
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
        RETURNING updated_at`
        
    err := r.db.QueryRow(
        query,
        tag.ID,
        tag.Name,
        tag.Colour,
    ).Scan(&tag.UpdatedAt)

    if err != nil {
        return fmt.Errorf("error updating tag: %v", err)
    }

    return nil
}

func (r *Repository) AssignTagsToItems(tagIDs []int, itemIDs []int) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    insertQuery := `
        INSERT INTO item_tag (tag_id, item_id)
        SELECT t.id, i.id
        FROM unnest($1::int[]) AS t(id)
        CROSS JOIN unnest($2::int[]) AS i(id)
        ON CONFLICT (tag_id, item_id) DO NOTHING
    `
    
    _, err = tx.Exec(insertQuery, pq.Array(tagIDs), pq.Array(itemIDs))
    if err != nil {
        return fmt.Errorf("error assigning tags: %v", err)
    }

    return tx.Commit()
}

func (r *Repository) Delete(id int) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    itemTagQuery := `DELETE FROM item_tag WHERE tag_id = $1`
    _, err = tx.Exec(itemTagQuery, id)
    if err != nil {
        return fmt.Errorf("error removing item-tag associations: %v", err)
    }

    tagQuery := `DELETE FROM tag WHERE id = $1`
    result, err := tx.Exec(tagQuery, id)
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

    return tx.Commit()
}