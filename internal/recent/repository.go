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

func (r *Repository) GetRecentEntities(userID int, limit int) (*Response, error) {
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

    // Containers
    containerQuery := `
        SELECT id, name, created_at 
        FROM container 
        WHERE user_id = $1 
        ORDER BY created_at DESC 
        LIMIT $2
    `
    containerRows, err := tx.Query(containerQuery, userID, limit)
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
    if err = containerRows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating container rows: %v", err)
    }

    containerCountQuery := `
        SELECT COUNT(*) 
        FROM container 
        WHERE user_id = $1
    `
    if err := tx.QueryRow(containerCountQuery, userID).Scan(&response.Containers.Total); err != nil {
        return nil, fmt.Errorf("failed to get container count: %v", err)
    }

    // Items
    itemQuery := `
        SELECT i.id, i.name, i.created_at 
        FROM item i
        JOIN container c ON i.container_id = c.id
        WHERE c.user_id = $1 
        ORDER BY i.created_at DESC 
        LIMIT $2
    `
    itemRows, err := tx.Query(itemQuery, userID, limit)
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
    if err = itemRows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating item rows: %v", err)
    }

    itemCountQuery := `
        SELECT COUNT(i.*) 
        FROM item i
        JOIN container c ON i.container_id = c.id
        WHERE c.user_id = $1
    `
    if err := tx.QueryRow(itemCountQuery, userID).Scan(&response.Items.Total); err != nil {
        return nil, fmt.Errorf("failed to get item count: %v", err)
    }

    // Tags
    tagQuery := `
        SELECT DISTINCT ON (t.id)
        t.id, t.name, t.created_at 
        FROM tag t
        JOIN item_tag it ON t.id = it.tag_id
        JOIN item i ON it.item_id = i.id
        JOIN container c ON i.container_id = c.id
        WHERE c.user_id = $1 
        ORDER BY t.id, t.created_at DESC 
        LIMIT $2
    `
    tagRows, err := tx.Query(tagQuery, userID, limit)
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
    if err = tagRows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating tag rows: %v", err)
    }

    tagCountQuery := `
        SELECT COUNT(DISTINCT t.id)
        FROM tag t
        JOIN item_tag it ON t.id = it.tag_id
        JOIN item i ON it.item_id = i.id
        JOIN container c ON i.container_id = c.id
        WHERE c.user_id = $1
    `
    if err := tx.QueryRow(tagCountQuery, userID).Scan(&response.Tags.Total); err != nil {
        return nil, fmt.Errorf("failed to get tag count: %v", err)
    }

    // Workspaces will be implemented later
    response.Workspaces.Total = 0

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %v", err)
    }

    return response, nil
}