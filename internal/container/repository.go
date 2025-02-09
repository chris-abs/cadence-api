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
            INSERT INTO item (name, description, quantity, container_id, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6)
            RETURNING id`

        for _, itemReq := range itemRequests {
            var itemID int
            err = tx.QueryRow(
                itemQuery,
                itemReq.Name,
                itemReq.Description,
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
               c.location, c.user_id, c.workspace_id, c.created_at, c.updated_at,
               w.id, w.name, w.description, w.user_id, w.created_at, w.updated_at
        FROM container c
        LEFT JOIN workspace w ON c.workspace_id = w.id
        WHERE c.id = $1`

    container := new(models.Container)
    var workspaceID sql.NullInt64
    var wsFields struct {
        ID          sql.NullInt64
        Name        sql.NullString
        Description sql.NullString
        UserID      sql.NullInt64
        CreatedAt   sql.NullTime
        UpdatedAt   sql.NullTime
    }

    err := r.db.QueryRow(containerQuery, id).Scan(
        &container.ID, &container.Name, &container.QRCode,
        &container.QRCodeImage, &container.Number, &container.Location,
        &container.UserID, &workspaceID, &container.CreatedAt, &container.UpdatedAt,
        &wsFields.ID, &wsFields.Name, &wsFields.Description,
        &wsFields.UserID, &wsFields.CreatedAt, &wsFields.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("container not found")
    }
    if err != nil {
        return nil, err
    }

    if workspaceID.Valid && wsFields.ID.Valid {
        wsID := int(workspaceID.Int64)
        container.WorkspaceID = &wsID
        container.Workspace = &models.Workspace{
            ID:          int(wsFields.ID.Int64),
            Name:        wsFields.Name.String,
            Description: wsFields.Description.String,
            UserID:      int(wsFields.UserID.Int64),
            CreatedAt:   wsFields.CreatedAt.Time,
            UpdatedAt:   wsFields.UpdatedAt.Time,
        }
    }

    itemsQuery := `
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
               i.container_id, i.created_at, i.updated_at,
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
        LEFT JOIN item_tag it ON i.id = it.item_id
        LEFT JOIN tag t ON it.tag_id = t.id
        WHERE i.container_id = $1
        GROUP BY i.id, i.name, i.description, i.quantity, 
                 i.container_id, i.created_at, i.updated_at,
                 img.images`

    rows, err := r.db.Query(itemsQuery, id)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    container.Items = make([]models.Item, 0)
    for rows.Next() {
        var item models.Item
        var imagesJSON, tagsJSON []byte

        err := rows.Scan(
            &item.ID, &item.Name, &item.Description,
            &item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
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

func (r *Repository) GetByUserID(userID int) ([]*models.Container, error) {
    query := `
        SELECT c.id, c.name, c.qr_code, c.qr_code_image, c.number, 
               c.location, c.user_id, c.workspace_id, c.created_at, c.updated_at,
               w.id, w.name, w.description, w.user_id, w.created_at, w.updated_at
        FROM container c
        LEFT JOIN workspace w ON c.workspace_id = w.id
        WHERE c.user_id = $1
        ORDER BY c.created_at DESC`

    rows, err := r.db.Query(query, userID)
    if err != nil {
        return nil, fmt.Errorf("error querying containers: %v", err)
    }
    defer rows.Close()

    var containers []*models.Container
    for rows.Next() {
        container := new(models.Container)
        var workspaceID sql.NullInt64
        var wsFields struct {
            ID          sql.NullInt64
            Name        sql.NullString
            Description sql.NullString
            UserID      sql.NullInt64
            CreatedAt   sql.NullTime
            UpdatedAt   sql.NullTime
        }

        err := rows.Scan(
            &container.ID, &container.Name, &container.QRCode,
            &container.QRCodeImage, &container.Number, &container.Location,
            &container.UserID, &workspaceID, &container.CreatedAt, &container.UpdatedAt,
            &wsFields.ID, &wsFields.Name, &wsFields.Description,
            &wsFields.UserID, &wsFields.CreatedAt, &wsFields.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning container: %v", err)
        }

        if workspaceID.Valid && wsFields.ID.Valid {
            wsID := int(workspaceID.Int64)
            container.WorkspaceID = &wsID
            container.Workspace = &models.Workspace{
                ID:          int(wsFields.ID.Int64),
                Name:        wsFields.Name.String,
                Description: wsFields.Description.String,
                UserID:      int(wsFields.UserID.Int64),
                CreatedAt:   wsFields.CreatedAt.Time,
                UpdatedAt:   wsFields.UpdatedAt.Time,
            }
        }

        itemsQuery := `
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
                   i.container_id, i.created_at, i.updated_at,
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
            LEFT JOIN item_tag it ON i.id = it.item_id
            LEFT JOIN tag t ON it.tag_id = t.id
            WHERE i.container_id = $1
            GROUP BY i.id, i.name, i.description, i.quantity, 
                     i.container_id, i.created_at, i.updated_at,
                     img.images`

        itemRows, err := r.db.Query(itemsQuery, container.ID)
        if err != nil {
            return nil, fmt.Errorf("error querying items: %v", err)
        }

        container.Items = make([]models.Item, 0)
        func() {
            defer itemRows.Close()
            for itemRows.Next() {
                var item models.Item
                var imagesJSON, tagsJSON []byte

                err := itemRows.Scan(
                    &item.ID, &item.Name, &item.Description,
                    &item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
                    &imagesJSON, &tagsJSON,
                )
                if err != nil {
                    return
                }

                if err := json.Unmarshal(imagesJSON, &item.Images); err != nil {
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
    return r.GetByQRWithItems(qrCode, true)
}

func (r *Repository) GetByQRWithItems(qrCode string, includeItems bool) (*models.Container, error) {
    query := `
        SELECT c.id, c.name, c.qr_code, c.qr_code_image, c.number, 
               c.location, c.user_id, c.workspace_id, c.created_at, c.updated_at,
               w.id, w.name, w.description, w.user_id, w.created_at, w.updated_at
        FROM container c
        LEFT JOIN workspace w ON c.workspace_id = w.id
        WHERE c.qr_code = $1`

    container := new(models.Container)
    var workspaceID sql.NullInt64
    var workspace models.Workspace

    err := r.db.QueryRow(query, qrCode).Scan(
        &container.ID, &container.Name, &container.QRCode,
        &container.QRCodeImage, &container.Number, &container.Location,
        &container.UserID, &workspaceID, &container.CreatedAt, &container.UpdatedAt,
        &workspace.ID, &workspace.Name, &workspace.Description,
        &workspace.UserID, &workspace.CreatedAt, &workspace.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("container not found")
    }
    if err != nil {
        return nil, err
    }

    if workspaceID.Valid {
        wsID := int(workspaceID.Int64)
        container.WorkspaceID = &wsID
        container.Workspace = &workspace
    }

    if includeItems {
        itemsQuery := `
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
                   i.container_id, i.created_at, i.updated_at,
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
            LEFT JOIN item_tag it ON i.id = it.item_id
            LEFT JOIN tag t ON it.tag_id = t.id
            WHERE i.container_id = $1
            GROUP BY i.id, i.name, i.description, i.quantity, 
                     i.container_id, i.created_at, i.updated_at,
                     img.images`

        rows, err := r.db.Query(itemsQuery, container.ID)
        if err != nil {
            return nil, err
        }
        defer rows.Close()

        container.Items = make([]models.Item, 0)
        for rows.Next() {
            var item models.Item
            var imagesJSON, tagsJSON []byte

            err := rows.Scan(
                &item.ID, &item.Name, &item.Description,
                &item.Quantity, &item.ContainerID, &item.CreatedAt, &item.UpdatedAt,
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
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    itemQuery := `
        UPDATE item 
        SET container_id = NULL, updated_at = $2
        WHERE container_id = $1`

    _, err = tx.Exec(itemQuery, id, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error updating items: %v", err)
    }

    workspaceQuery := `
        UPDATE container
        SET workspace_id = NULL, updated_at = $2
        WHERE id = $1`

    _, err = tx.Exec(workspaceQuery, id, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error removing workspace reference: %v", err)
    }

    containerQuery := `DELETE FROM container WHERE id = $1`
    result, err := tx.Exec(containerQuery, id)
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