package tag

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/storage/entities"
	"github.com/lib/pq"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) Create(tag *entities.Tag) error {
    query := `
        INSERT INTO tag (name, description, colour, profile_id, family_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`

    err := r.db.QueryRow(
        query,
        tag.Name,
        tag.Description,
        tag.Colour,
        tag.ProfileID,
        tag.FamilyID,
        tag.CreatedAt,
        tag.UpdatedAt,
    ).Scan(&tag.ID)

    if err != nil {
        return fmt.Errorf("error creating tag: %v", err)
    }

    return nil
}

func (r *Repository) GetByID(id int, familyID int) (*entities.Tag, error) {
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
        SELECT t.id, t.name, t.colour, t.description, t.profile_id, t.family_id, t.created_at, t.updated_at,
               COALESCE(
                   jsonb_agg(
                       DISTINCT jsonb_build_object(
                           'id', i.id,
                           'name', i.name,
                           'description', i.description,
                           'images', COALESCE(img.images, '[]'::jsonb),
                           'quantity', i.quantity,
                           'containerId', i.container_id,
                           'familyId', i.family_id,
                           'container', CASE 
                               WHEN c.id IS NOT NULL THEN
                                   jsonb_build_object(
                                       'id', c.id,
                                       'name', c.name,
                                       'qrCode', c.qr_code,
                                       'qrCodeImage', c.qr_code_image,
                                       'number', c.number,
                                       'location', c.location,
                                       'familyId', c.family_id,
                                       'workspaceId', c.workspace_id,
                                       'workspace', CASE 
                                           WHEN w.id IS NOT NULL THEN
                                               jsonb_build_object(
                                                   'id', w.id,
                                                   'name', w.name,
                                                   'description', w.description,
                                                   'familyId', w.family_id,
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
                           'updatedAt', i.updated_at
                       )
                   ) FILTER (WHERE i.id IS NOT NULL),
                   '[]'
               ) as items
        FROM tag t
        LEFT JOIN item_tag it ON t.id = it.tag_id
        LEFT JOIN item i ON it.item_id = i.id AND i.family_id = t.family_id AND i.is_deleted = false
        LEFT JOIN item_images img ON i.id = img.item_id AND img.is_deleted = false
        LEFT JOIN container c ON i.container_id = c.id AND c.family_id = t.family_id AND c.is_deleted = false
        LEFT JOIN workspace w ON c.workspace_id = w.id AND w.family_id = t.family_id AND w.is_deleted = false
        WHERE t.id = $1 AND t.family_id = $2 AND t.is_deleted = false
        GROUP BY t.id, t.name, t.colour, t.family_id, t.created_at, t.updated_at`

    tag := new(entities.Tag)
    var itemsJSON []byte

    err := r.db.QueryRow(query, id, familyID).Scan(
        &tag.ID, &tag.Name, &tag.Colour, &tag.Description, 
        &tag.ProfileID, &tag.FamilyID,
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

func (r *Repository) GetByFamilyID(familyID int) ([]*entities.Tag, error) {
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
            t.description, t.profile_id, t.family_id, t.created_at, t.updated_at,
               COALESCE(
                   jsonb_agg(
                       DISTINCT jsonb_build_object(
                           'id', i.id,
                           'name', i.name,
                           'description', i.description,
                           'images', COALESCE(img.images, '[]'::jsonb),
                           'quantity', i.quantity,
                           'containerId', i.container_id,
                           'familyId', i.family_id,
                           'container', CASE 
                               WHEN c.id IS NOT NULL THEN
                                   jsonb_build_object(
                                       'id', c.id,
                                       'name', c.name,
                                       'qrCode', c.qr_code,
                                       'qrCodeImage', c.qr_code_image,
                                       'number', c.number,
                                       'location', c.location,
                                       'familyId', c.family_id,
                                       'workspaceId', c.workspace_id,
                                       'workspace', CASE 
                                           WHEN w.id IS NOT NULL THEN
                                               jsonb_build_object(
                                                   'id', w.id,
                                                   'name', w.name,
                                                   'description', w.description,
                                                   'familyId', w.family_id,
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
                           'updatedAt', i.updated_at
                       )
                   ) FILTER (WHERE i.id IS NOT NULL),
                   '[]'
               ) as items
        FROM tag t
        LEFT JOIN item_tag it ON t.id = it.tag_id
        LEFT JOIN item i ON it.item_id = i.id AND i.family_id = t.family_id AND i.is_deleted = false
        LEFT JOIN item_images img ON i.id = img.item_id AND img.is_deleted = false
        LEFT JOIN container c ON i.container_id = c.id AND c.family_id = t.family_id AND c.is_deleted = false
        LEFT JOIN workspace w ON c.workspace_id = w.id AND w.family_id = t.family_id AND w.is_deleted = false
        WHERE t.family_id = $1 AND t.is_deleted = false
        GROUP BY t.id, t.name, t.colour, t.family_id, t.created_at, t.updated_at
        ORDER BY t.name ASC`

    rows, err := r.db.Query(query, familyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tags []*entities.Tag
    for rows.Next() {
        tag := new(entities.Tag)
        var itemsJSON []byte

        err := rows.Scan(
            &tag.ID, &tag.Name, &tag.Colour, &tag.Description,
            &tag.ProfileID, &tag.FamilyID,
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

func (r *Repository) Update(tag *entities.Tag) error {
    query := `
        UPDATE tag
        SET name = $2, 
            description = $3,
            colour = $4, 
            profile_id = $5,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND family_id = $6 AND is_deleted = false
        RETURNING updated_at`
            
    err := r.db.QueryRow(
        query,
        tag.ID,
        tag.Name,
        tag.Description,
        tag.Colour,
        tag.ProfileID,
        tag.FamilyID,
    ).Scan(&tag.UpdatedAt)

    if err != nil {
        return fmt.Errorf("error updating tag: %v", err)
    }

    return nil
}

func (r *Repository) AssignTagsToItems(familyID int, tagIDs []int, itemIDs []int) error {
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
        WHERE EXISTS (
            SELECT 1 FROM tag WHERE id = t.id AND family_id = $3
        ) AND EXISTS (
            SELECT 1 FROM item WHERE id = i.id AND family_id = $3
        )
        ON CONFLICT (tag_id, item_id) DO NOTHING
    `
    
    _, err = tx.Exec(insertQuery, pq.Array(tagIDs), pq.Array(itemIDs), familyID)
    if err != nil {
        return fmt.Errorf("error assigning tags: %v", err)
    }

    return tx.Commit()
}

func (r *Repository) Delete(id int, familyID int, deletedBy int) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    itemTagQuery := `
        DELETE FROM item_tag
        WHERE tag_id = $1 AND is_deleted = false
        AND EXISTS (
            SELECT 1 FROM tag
            WHERE id = $1 AND family_id = $2 AND is_deleted = false
        )`
    
    _, err = tx.Exec(itemTagQuery, id, familyID)
    if err != nil {
        return fmt.Errorf("error removing item-tag associations: %v", err)
    }

    tagQuery := `
        UPDATE tag 
        SET is_deleted = true, deleted_at = $3, deleted_by = $4, updated_at = $3
        WHERE id = $1 AND family_id = $2 AND is_deleted = false`
        
    result, err := tx.Exec(tagQuery, id, familyID, time.Now().UTC(), deletedBy)
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

func (r *Repository) RestoreDeleted(id int, familyID int) error {
    query := `
        UPDATE tag
        SET is_deleted = false, deleted_at = NULL, deleted_by = NULL, updated_at = $3
        WHERE id = $1 AND family_id = $2 AND is_deleted = true`
    
    result, err := r.db.Exec(query, id, familyID, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error restoring tag: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking restore result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("tag not found or not deleted")
    }

    return nil
}