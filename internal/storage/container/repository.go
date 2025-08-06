package container

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/models"
	"github.com/chrisabs/cadence/internal/storage/entities"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) Create(container *entities.Container, itemRequests []CreateItemRequest) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    containerQuery := `
        INSERT INTO container (
            id, name, description, qr_code, qr_code_image, number, 
            location, profile_id, family_id, workspace_id, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

    err = tx.QueryRow(
        containerQuery,
        container.ID,
        container.Name,
        container.Description,
        container.QRCode,
        container.QRCodeImage,
        container.Number,
        container.Location,
        container.ProfileID,
        container.FamilyID,
        container.WorkspaceID,
        container.CreatedAt,
        container.UpdatedAt,
    ).Scan()

    if err != nil && err != sql.ErrNoRows {
        return fmt.Errorf("error creating container: %v", err)
    }

    if len(itemRequests) > 0 {
        itemQuery := `
            INSERT INTO item (
                id, name, description, quantity, container_id, 
                profile_id, family_id, created_at, updated_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

        for _, itemReq := range itemRequests {
            itemID := models.NewItemID()
            _, err = tx.Exec(
                itemQuery,
                itemID,
                itemReq.Name,
                itemReq.Description,
                itemReq.Quantity,
                container.ID,
                container.ProfileID,
                container.FamilyID,
                time.Now().UTC(),
                time.Now().UTC(),
            )

            if err != nil {
                return fmt.Errorf("error creating item: %v", err)
            }
        }
    }

    return tx.Commit()
}

func (r *Repository) GetByID(id models.ContainerID, familyID models.FamilyID) (*entities.Container, error) {
    containerQuery := `
        SELECT c.id, c.name, c.description, c.qr_code, c.qr_code_image, c.number, 
            c.location, c.profile_id, c.family_id, c.workspace_id, c.created_at, c.updated_at,
            w.id, w.name, w.description, w.profile_id, w.family_id, w.created_at, w.updated_at
        FROM container c
        LEFT JOIN workspace w ON c.workspace_id = w.id AND w.family_id = c.family_id AND w.is_deleted = false
        WHERE c.id = $1 AND c.family_id = $2 AND c.is_deleted = false`

    container := new(entities.Container)
    var workspaceID sql.NullString
    var wsFields struct {
        ID          sql.NullString
        Name        sql.NullString
        Description sql.NullString
        ProfileID   sql.NullString
        FamilyID    sql.NullString
        CreatedAt   sql.NullTime
        UpdatedAt   sql.NullTime
    }

    err := r.db.QueryRow(containerQuery, id, familyID).Scan(
        &container.ID, &container.Name, &container.Description, &container.QRCode,
        &container.QRCodeImage, &container.Number, &container.Location,
        &container.ProfileID,
        &container.FamilyID, &workspaceID,
        &container.CreatedAt, &container.UpdatedAt,
        &wsFields.ID, &wsFields.Name, &wsFields.Description,
        &wsFields.ProfileID, &wsFields.FamilyID, &wsFields.CreatedAt, &wsFields.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("container not found")
    }
    if err != nil {
        return nil, err
    }

    if workspaceID.Valid && wsFields.ID.Valid {
        wsID := models.WorkspaceID(workspaceID.String)
        container.WorkspaceID = &wsID
        container.Workspace = &entities.Workspace{
            ID:          models.WorkspaceID(wsFields.ID.String),
            Name:        wsFields.Name.String,
            Description: wsFields.Description.String,
            ProfileID:   models.ProfileID(wsFields.ProfileID.String),
            FamilyID:    models.FamilyID(wsFields.FamilyID.String),
            CreatedAt:   wsFields.CreatedAt.Time,
            UpdatedAt:   wsFields.UpdatedAt.Time,
        }
    }

    itemsQuery := `
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
            WHERE is_deleted = false 
            GROUP BY item_id
        )
        SELECT i.id, i.name, i.description, i.quantity, 
               i.container_id, i.profile_id, i.family_id, i.created_at, i.updated_at,
               COALESCE(img.images, '[]'::jsonb) as images,
               COALESCE(
                   jsonb_agg(
                       DISTINCT jsonb_build_object(
                           'id', t.id,
                           'name', t.name,
                           'colour', t.colour,
                           'createdAt', t.created_at,
                           'updatedAt', t.updated_at 
                       )
                   ) FILTER (WHERE t.id IS NOT NULL),
                   '[]'
               ) as tags
        FROM item i
        LEFT JOIN item_images img ON i.id = img.item_id
        LEFT JOIN item_tag it ON i.id = it.item_id AND it.is_deleted = false
        LEFT JOIN tag t ON it.tag_id = t.id AND t.family_id = i.family_id AND t.is_deleted = false
        WHERE i.container_id = $1 AND i.family_id = $2 AND i.is_deleted = false
        GROUP BY i.id, i.name, i.description, i.quantity, 
                 i.container_id, i.profile_id, i.family_id, i.created_at, i.updated_at,
                 img.images`

    rows, err := r.db.Query(itemsQuery, id, familyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    container.Items = make([]entities.Item, 0)
    for rows.Next() {
        var item entities.Item
        var imagesJSON, tagsJSON []byte

        err := rows.Scan(
            &item.ID, &item.Name, &item.Description,
            &item.Quantity, &item.ContainerID, &item.ProfileID, &item.FamilyID,
            &item.CreatedAt, &item.UpdatedAt,
            &imagesJSON, &tagsJSON,
        )
        if err != nil {
            return nil, err
        }

        if err := json.Unmarshal(imagesJSON, &item.Images); err != nil {
            return nil, fmt.Errorf("error parsing images: %v", err)
        }

        if err := json.Unmarshal(tagsJSON, &item.Tags); err != nil {
            return nil, fmt.Errorf("error parsing tags: %v", err)
        }

        container.Items = append(container.Items, item)
    }

    return container, nil
}

func (r *Repository) GetByFamilyID(familyID models.FamilyID) ([]*entities.Container, error) {
    query := `
        SELECT c.id, c.name, c.description, c.qr_code, c.qr_code_image, c.number, 
            c.location, c.profile_id, c.family_id, c.workspace_id, c.created_at, c.updated_at,
            w.id, w.name, w.description, w.profile_id, w.family_id, w.created_at, w.updated_at
        FROM container c
        LEFT JOIN workspace w ON c.workspace_id = w.id AND w.family_id = c.family_id AND w.is_deleted = false
        WHERE c.family_id = $1 AND c.is_deleted = false
        ORDER BY c.created_at DESC`

    rows, err := r.db.Query(query, familyID)
    if err != nil {
        return nil, fmt.Errorf("error querying containers: %v", err)
    }
    defer rows.Close()

    var containers []*entities.Container
    for rows.Next() {
        container := new(entities.Container)
        var workspaceID sql.NullString
        var wsFields struct {
            ID          sql.NullString
            Name        sql.NullString
            Description sql.NullString
            ProfileID   sql.NullString
            FamilyID    sql.NullString
            CreatedAt   sql.NullTime
            UpdatedAt   sql.NullTime
        }

        err := rows.Scan(
            &container.ID, &container.Name, &container.Description, &container.QRCode,
            &container.QRCodeImage, &container.Number, &container.Location,
            &container.ProfileID,
            &container.FamilyID, &workspaceID,
            &container.CreatedAt, &container.UpdatedAt,
            &wsFields.ID, &wsFields.Name, &wsFields.Description,
            &wsFields.ProfileID, &wsFields.FamilyID, &wsFields.CreatedAt, &wsFields.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning container: %v", err)
        }

        if workspaceID.Valid && wsFields.ID.Valid {
            wsID := models.WorkspaceID(workspaceID.String)
            container.WorkspaceID = &wsID
            container.Workspace = &entities.Workspace{
                ID:          models.WorkspaceID(wsFields.ID.String),
                Name:        wsFields.Name.String,
                Description: wsFields.Description.String,
                ProfileID:   models.ProfileID(wsFields.ProfileID.String),
                FamilyID:    models.FamilyID(wsFields.FamilyID.String),
                CreatedAt:   wsFields.CreatedAt.Time,
                UpdatedAt:   wsFields.UpdatedAt.Time,
            }
        }

        container.Items = make([]entities.Item, 0)
        containers = append(containers, container)
    }

    return containers, nil
}

func (r *Repository) GetByQR(qrCode string, familyID models.FamilyID) (*entities.Container, error) {
    return r.getContainerByField("qr_code", qrCode, familyID)
}

func (r *Repository) getContainerByField(field, value string, familyID models.FamilyID) (*entities.Container, error) {
    query := fmt.Sprintf(`
        SELECT c.id, c.name, c.description, c.qr_code, c.qr_code_image, c.number, 
            c.location, c.profile_id, c.family_id, c.workspace_id, c.created_at, c.updated_at,
            w.id, w.name, w.description, w.profile_id, w.family_id, w.created_at, w.updated_at
        FROM container c
        LEFT JOIN workspace w ON c.workspace_id = w.id AND w.family_id = c.family_id AND w.is_deleted = false
        WHERE c.%s = $1 AND c.family_id = $2 AND c.is_deleted = false`, field)

    container := new(entities.Container)
    var workspaceID sql.NullString
    var wsFields struct {
        ID          sql.NullString
        Name        sql.NullString
        Description sql.NullString
        ProfileID   sql.NullString
        FamilyID    sql.NullString
        CreatedAt   sql.NullTime
        UpdatedAt   sql.NullTime
    }

    err := r.db.QueryRow(query, value, familyID).Scan(
        &container.ID, &container.Name, &container.Description, &container.QRCode,
        &container.QRCodeImage, &container.Number, &container.Location,
        &container.ProfileID,
        &container.FamilyID, &workspaceID,
        &container.CreatedAt, &container.UpdatedAt,
        &wsFields.ID, &wsFields.Name, &wsFields.Description,
        &wsFields.ProfileID, &wsFields.FamilyID, &wsFields.CreatedAt, &wsFields.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("container not found")
    }
    if err != nil {
        return nil, err
    }

    if workspaceID.Valid && wsFields.ID.Valid {
        wsID := models.WorkspaceID(workspaceID.String)
        container.WorkspaceID = &wsID
        container.Workspace = &entities.Workspace{
            ID:          models.WorkspaceID(wsFields.ID.String),
            Name:        wsFields.Name.String,
            Description: wsFields.Description.String,
            ProfileID:   models.ProfileID(wsFields.ProfileID.String),
            FamilyID:    models.FamilyID(wsFields.FamilyID.String),
            CreatedAt:   wsFields.CreatedAt.Time,
            UpdatedAt:   wsFields.UpdatedAt.Time,
        }
    }

    return container, nil
}

func (r *Repository) Update(container *entities.Container) error {
    query := `
        UPDATE container
        SET name = $2, description = $3, location = $4, workspace_id = $5, updated_at = $6
        WHERE id = $1 AND family_id = $7 AND is_deleted = false`

    var workspaceID sql.NullString
    if container.WorkspaceID != nil {
        workspaceID = sql.NullString{String: container.WorkspaceID.String(), Valid: true}
    }

    result, err := r.db.Exec(
        query,
        container.ID,
        container.Name,
        container.Description,
        container.Location,
        workspaceID,
        time.Now().UTC(),
        container.FamilyID,
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

func (r *Repository) Delete(id models.ContainerID, familyID models.FamilyID, deletedBy models.ProfileID) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    itemQuery := `
        UPDATE item 
        SET container_id = NULL, updated_at = $3
        WHERE container_id = $1 AND family_id = $2 AND is_deleted = false`

    _, err = tx.Exec(itemQuery, id, familyID, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error updating items: %v", err)
    }

    containerQuery := `
        UPDATE container 
        SET is_deleted = true, deleted_at = $3, deleted_by = $4, updated_at = $3
        WHERE id = $1 AND family_id = $2 AND is_deleted = false`

    result, err := tx.Exec(containerQuery, id, familyID, time.Now().UTC(), deletedBy)
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

    return tx.Commit()
}

func (r *Repository) RestoreDeleted(id models.ContainerID, familyID models.FamilyID) error {
    query := `
        UPDATE container
        SET is_deleted = false, deleted_at = NULL, deleted_by = NULL, updated_at = $3
        WHERE id = $1 AND family_id = $2 AND is_deleted = true`

    result, err := r.db.Exec(query, id, familyID, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error restoring container: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking restore result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("container not found or not deleted")
    }

    return nil
}