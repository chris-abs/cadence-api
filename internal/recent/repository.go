package recent

import (
	"database/sql"
	"fmt"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) GetRecentEntities(familyID int, limit int) (*Response, error) {
    tx, err := r.db.Begin()
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %v", err)
    }
    defer tx.Rollback()

    response := &Response{
        Workspaces: EntityStats{Recent: make([]EntityPreview, 0)},
        Containers: EntityStats{Recent: make([]EntityPreview, 0)},
        Items:      EntityStats{Recent: make([]EntityPreview, 0)},
        Tags:       EntityStats{Recent: make([]EntityPreview, 0)},
    }

    containerCountQuery := `
        SELECT COUNT(*) 
        FROM container 
        WHERE family_id = $1
    `
    if err := tx.QueryRow(containerCountQuery, familyID).Scan(&response.Containers.Total); err != nil {
        return nil, fmt.Errorf("failed to get container count: %v", err)
    }

    containerQuery := `
        SELECT id, name, created_at 
        FROM container 
        WHERE family_id = $1 
        ORDER BY created_at DESC 
        LIMIT $2
    `
    containerRows, err := tx.Query(containerQuery, familyID, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch recent containers: %v", err)
    }
    defer containerRows.Close()

    for containerRows.Next() {
        var preview EntityPreview
        if err := containerRows.Scan(&preview.ID, &preview.Name, &preview.CreatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan container row: %v", err)
        }
        response.Containers.Recent = append(response.Containers.Recent, preview)
    }

    itemCountQuery := `
        SELECT COUNT(DISTINCT i.id) 
        FROM item i
        WHERE i.family_id = $1
    `
    if err := tx.QueryRow(itemCountQuery, familyID).Scan(&response.Items.Total); err != nil {
        return nil, fmt.Errorf("failed to get item count: %v", err)
    }

    itemQuery := `
        SELECT i.id, i.name, i.created_at 
        FROM item i
        WHERE i.family_id = $1
        ORDER BY i.created_at DESC 
        LIMIT $2
    `

    itemRows, err := tx.Query(itemQuery, familyID, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch recent items: %v", err)
    }
    defer itemRows.Close()

    for itemRows.Next() {
        var preview EntityPreview
        if err := itemRows.Scan(&preview.ID, &preview.Name, &preview.CreatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan item row: %v", err)
        }
        response.Items.Recent = append(response.Items.Recent, preview)
    }

    tagCountQuery := `
        SELECT COUNT(DISTINCT t.id)
        FROM tag t
        WHERE t.family_id = $1
    `
    if err := tx.QueryRow(tagCountQuery, familyID).Scan(&response.Tags.Total); err != nil {
        return nil, fmt.Errorf("failed to get tag count: %v", err)
    }

    tagQuery := `
        SELECT DISTINCT t.id, t.name, t.created_at 
        FROM tag t
        WHERE t.family_id = $1
        ORDER BY t.created_at DESC 
        LIMIT $2
    `
    tagRows, err := tx.Query(tagQuery, familyID, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch recent tags: %v", err)
    }
    defer tagRows.Close()

    for tagRows.Next() {
        var preview EntityPreview
        if err := tagRows.Scan(&preview.ID, &preview.Name, &preview.CreatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan tag row: %v", err)
        }
        response.Tags.Recent = append(response.Tags.Recent, preview)
    }

    workspaceCountQuery := `
        SELECT COUNT(*) 
        FROM workspace 
        WHERE family_id = $1
    `
    if err := tx.QueryRow(workspaceCountQuery, familyID).Scan(&response.Workspaces.Total); err != nil {
        return nil, fmt.Errorf("failed to get workspace count: %v", err)
    }

    workspaceQuery := `
        SELECT id, name, created_at 
        FROM workspace 
        WHERE family_id = $1
        ORDER BY created_at DESC 
        LIMIT $2
    `
    workspaceRows, err := tx.Query(workspaceQuery, familyID, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch recent workspaces: %v", err)
    }
    defer workspaceRows.Close()

    for workspaceRows.Next() {
        var preview EntityPreview
        if err := workspaceRows.Scan(&preview.ID, &preview.Name, &preview.CreatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan workspace row: %v", err)
        }
        response.Workspaces.Recent = append(response.Workspaces.Recent, preview)
    }

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %v", err)
    }

    return response, nil
}