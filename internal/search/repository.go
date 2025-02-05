package search

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) Search(query string, userID int) (*SearchResponse, error) {
    sqlQuery := `
    WITH workspace_matches AS (
        SELECT 
            'workspace' as type,
            id,
            name,
            description,
            ts_rank(to_tsvector('english', name || ' ' || COALESCE(description, '')), 
                   plainto_tsquery('english', $1)) as rank,
            NULL as container_name,
            name as workspace_name,
            NULL as colour            -- Explicitly include colour column
        FROM workspace 
        WHERE 
            user_id = $2 AND
            to_tsvector('english', name || ' ' || COALESCE(description, '')) @@ 
            plainto_tsquery('english', $1)
    ),
    container_matches AS (
        SELECT 
            'container' as type,
            c.id,
            c.name,
            '' as description,
            ts_rank(to_tsvector('english', c.name), plainto_tsquery('english', $1)) as rank,
            NULL as container_name,
            w.name as workspace_name,
            NULL as colour            -- Explicitly include colour column
        FROM container c
        LEFT JOIN workspace w ON c.workspace_id = w.id
        WHERE 
            c.user_id = $2 AND
            to_tsvector('english', c.name) @@ plainto_tsquery('english', $1)
    ),
    item_matches AS (
        SELECT 
            'item' as type,
            i.id,
            i.name,
            i.description,
            ts_rank(to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')), 
                   plainto_tsquery('english', $1)) as rank,
            c.name as container_name,
            w.name as workspace_name,
            NULL as colour            -- Explicitly include colour column
        FROM item i
        LEFT JOIN container c ON i.container_id = c.id
        LEFT JOIN workspace w ON c.workspace_id = w.id
        WHERE 
            (c.user_id = $2 OR i.container_id IS NULL) AND
            to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')) @@ 
            plainto_tsquery('english', $1)
    ),
    tag_matches AS (
        SELECT DISTINCT
            'tag' as type,
            t.id,
            t.name,
            '' as description,
            ts_rank(to_tsvector('english', t.name), plainto_tsquery('english', $1)) as rank,
            NULL as container_name,
            NULL as workspace_name,
            t.colour                  -- Include the actual tag colour
        FROM tag t
        LEFT JOIN item_tag it ON t.id = it.tag_id
        LEFT JOIN item i ON it.item_id = i.id
        LEFT JOIN container c ON i.container_id = c.id
        WHERE 
            (c.user_id = $2 OR i.container_id IS NULL) AND
            to_tsvector('english', t.name) @@ plainto_tsquery('english', $1)
    ),
    tagged_items AS (
        SELECT DISTINCT
            'tagged_item' as type,
            i.id,
            i.name,
            i.description,
            ts_rank(to_tsvector('english', t.name), plainto_tsquery('english', $1)) as rank,
            c.name as container_name,
            w.name as workspace_name,
            NULL as colour            -- Explicitly include colour column
        FROM item i
        INNER JOIN item_tag it ON i.id = it.item_id
        INNER JOIN tag t ON it.tag_id = t.id
        LEFT JOIN container c ON i.container_id = c.id
        LEFT JOIN workspace w ON c.workspace_id = w.id
        WHERE 
            (c.user_id = $2 OR i.container_id IS NULL) AND
            to_tsvector('english', t.name) @@ plainto_tsquery('english', $1) AND
            i.id NOT IN (SELECT id FROM item_matches)
    )
    SELECT type, id, name, description, rank, container_name, workspace_name, colour 
    FROM (
        SELECT * FROM workspace_matches
        UNION ALL
        SELECT * FROM container_matches
        UNION ALL
        SELECT * FROM item_matches
        UNION ALL
        SELECT * FROM tag_matches
        UNION ALL
        SELECT * FROM tagged_items
    ) combined_results
    ORDER BY rank DESC;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing search: %v", err)
    }
    defer rows.Close()

    response := &SearchResponse{
        Workspaces:  make([]SearchResult, 0),
        Containers:  make([]SearchResult, 0),
        Items:       make([]SearchResult, 0),
        Tags:        make([]SearchResult, 0),
        TaggedItems: make([]SearchResult, 0),
    }

    for rows.Next() {
        var result SearchResult
        var containerName, workspaceName, colour sql.NullString
        err := rows.Scan(
            &result.Type,
            &result.ID,
            &result.Name,
            &result.Description,
            &result.Rank,
            &containerName,
            &workspaceName,
            &colour,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning search result: %v", err)
        }
    
        if containerName.Valid {
            result.ContainerName = &containerName.String
        }
        if workspaceName.Valid {
            result.WorkspaceName = &workspaceName.String
        }

        if colour.Valid {
            result.Colour = &colour.String
        }

        switch result.Type {
        case "workspace":
            response.Workspaces = append(response.Workspaces, result)
        case "container":
            response.Containers = append(response.Containers, result)
        case "item":
            response.Items = append(response.Items, result)
        case "tag":
            response.Tags = append(response.Tags, result)
        case "tagged_item":
            response.TaggedItems = append(response.TaggedItems, result)
    }
}

    return response, nil
}

func (r *Repository) SearchWorkspaces(query string, userID int) (WorkspaceSearchResults, error) {
    sqlQuery := `
        SELECT 
            w.*,
            ts_rank(to_tsvector('english', w.name || ' ' || COALESCE(w.description, '')), 
                   plainto_tsquery('english', $1)) as rank,
            COALESCE(json_agg(
                DISTINCT jsonb_build_object(
                    'id', c.id,
                    'name', c.name,
                    'location', c.location,
                    'qrCode', c.qr_code,
                    'qrCodeImage', c.qr_code_image,
                    'number', c.number,
                    'userId', c.user_id,
                    'workspaceId', c.workspace_id,
                    'createdAt', c.created_at,
                    'updatedAt', c.updated_at
                )
            ) FILTER (WHERE c.id IS NOT NULL), '[]'::json) as containers
        FROM workspace w
        LEFT JOIN container c ON w.id = c.workspace_id
        WHERE 
            w.user_id = $2 AND
            to_tsvector('english', w.name || ' ' || COALESCE(w.description, '')) @@ 
            plainto_tsquery('english', $1)
        GROUP BY w.id
        ORDER BY rank DESC;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing workspace search: %v", err)
    }
    defer rows.Close()

    var results WorkspaceSearchResults
    for rows.Next() {
        var result WorkspaceSearchResult
        var containersJSON []byte
        
        err := rows.Scan(
            &result.ID,
            &result.Name,
            &result.Description,
            &result.UserID,
            &result.CreatedAt,
            &result.UpdatedAt,
            &result.Rank,
            &containersJSON,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning workspace search result: %v", err)
        }

        if len(containersJSON) > 0 {
            if err := json.Unmarshal(containersJSON, &result.Containers); err != nil {
                return nil, fmt.Errorf("error unmarshaling containers: %v", err)
            }
        }

        results = append(results, result)
    }

    return results, nil
}


func (r *Repository) SearchContainers(query string, userID int) (ContainerSearchResults, error) {
    sqlQuery := `
        SELECT 
            c.*,
            ts_rank(to_tsvector('english', c.name || ' ' || COALESCE(c.location, '')), 
                   plainto_tsquery('english', $1)) as rank,
            row_to_json(w.*) as workspace,
            COALESCE(json_agg(
                DISTINCT jsonb_build_object(
                    'id', i.id,
                    'name', i.name,
                    'description', i.description,
                    'quantity', i.quantity,
                    'containerId', i.container_id,
                    'createdAt', i.created_at,
                    'updatedAt', i.updated_at,
                    'tags', (
                        SELECT json_agg(
                            jsonb_build_object(
                                'id', t.id,
                                'name', t.name,
                                'colour', t.colour
                            )
                        )
                        FROM item_tag it
                        JOIN tag t ON it.tag_id = t.id
                        WHERE it.item_id = i.id
                    )
                )
            ) FILTER (WHERE i.id IS NOT NULL), '[]'::json) as items
        FROM container c
        LEFT JOIN workspace w ON c.workspace_id = w.id
        LEFT JOIN item i ON c.id = i.container_id
        WHERE 
            c.user_id = $2 AND
            to_tsvector('english', c.name || ' ' || COALESCE(c.location, '')) @@ 
            plainto_tsquery('english', $1)
        GROUP BY c.id, w.id
        ORDER BY rank DESC;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing container search: %v", err)
    }
    defer rows.Close()

    var results ContainerSearchResults
    for rows.Next() {
        var result ContainerSearchResult
        var workspaceJSON, itemsJSON []byte

        err := rows.Scan(
            &result.ID,
            &result.Name,
            &result.QRCode,
            &result.QRCodeImage,
            &result.Number,
            &result.Location,
            &result.UserID,
            &result.WorkspaceID,
            &result.CreatedAt,
            &result.UpdatedAt,
            &result.Rank,
            &workspaceJSON,
            &itemsJSON,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning container search result: %v", err)
        }

        if len(itemsJSON) > 0 {
            if err := json.Unmarshal(itemsJSON, &result.Items); err != nil {
                return nil, fmt.Errorf("error unmarshaling items: %v", err)
            }
        }

        results = append(results, result)
    }

    return results, nil
}

func (r *Repository) SearchItems(query string, userID int) (ItemSearchResults, error) {
    sqlQuery := `
        SELECT 
            i.*,
            ts_rank(to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')), 
                   plainto_tsquery('english', $1)) as rank,
            COALESCE(json_agg(
                DISTINCT jsonb_build_object(
                    'id', img.id,
                    'url', img.url,
                    'displayOrder', img.display_order,
                    'createdAt', img.created_at,
                    'updatedAt', img.updated_at
                )
            ) FILTER (WHERE img.id IS NOT NULL), '[]'::json) as images,
            COALESCE(json_agg(
                DISTINCT jsonb_build_object(
                    'id', t.id,
                    'name', t.name,
                    'colour', t.colour,
                    'createdAt', t.created_at,
                    'updatedAt', t.updated_at
                )
            ) FILTER (WHERE t.id IS NOT NULL), '[]'::json) as tags,
            row_to_json(c.*) as container
        FROM item i
        LEFT JOIN container c ON i.container_id = c.id
        LEFT JOIN item_image img ON i.id = img.item_id
        LEFT JOIN item_tag it ON i.id = it.item_id
        LEFT JOIN tag t ON it.tag_id = t.id
        WHERE 
            (c.user_id = $2 OR i.container_id IS NULL) AND
            to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')) @@ 
            plainto_tsquery('english', $1)
        GROUP BY i.id, c.id
        ORDER BY rank DESC;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing item search: %v", err)
    }
    defer rows.Close()

    var results ItemSearchResults
    for rows.Next() {
        var result ItemSearchResult
        var imagesJSON, tagsJSON, containerJSON []byte

        err := rows.Scan(
            &result.ID,
            &result.Name,
            &result.Description,
            &result.Quantity,
            &result.ContainerID,
            &result.CreatedAt,
            &result.UpdatedAt,
            &result.Rank,
            &imagesJSON,
            &tagsJSON,
            &containerJSON,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning item search result: %v", err)
        }

        if len(imagesJSON) > 0 {
            if err := json.Unmarshal(imagesJSON, &result.Images); err != nil {
                return nil, fmt.Errorf("error unmarshaling images: %v", err)
            }
        }

        if len(tagsJSON) > 0 {
            if err := json.Unmarshal(tagsJSON, &result.Tags); err != nil {
                return nil, fmt.Errorf("error unmarshaling tags: %v", err)
            }
        }

        if len(containerJSON) > 0 {
            if err := json.Unmarshal(containerJSON, &result.Container); err != nil {
                return nil, fmt.Errorf("error unmarshaling container: %v", err)
            }
        }

        results = append(results, result)
    }

    return results, nil
}

func (r *Repository) SearchTags(query string, userID int) (TagSearchResults, error) {
    sqlQuery := `
        SELECT 
            t.*,
            ts_rank(to_tsvector('english', t.name), plainto_tsquery('english', $1)) as rank,
            COALESCE(json_agg(
                DISTINCT jsonb_build_object(
                    'id', i.id,
                    'name', i.name,
                    'description', i.description,
                    'quantity', i.quantity,
                    'containerId', i.container_id,
                    'createdAt', i.created_at,
                    'updatedAt', i.updated_at,
                    'container', (
                        SELECT row_to_json(c.*)
                        FROM container c
                        WHERE c.id = i.container_id
                    ),
                    'images', (
                        SELECT json_agg(
                            jsonb_build_object(
                                'id', img.id,
                                'url', img.url,
                                'displayOrder', img.display_order
                            )
                        )
                        FROM item_image img
                        WHERE img.item_id = i.id
                    )
                )
            ) FILTER (WHERE i.id IS NOT NULL), '[]'::json) as items
        FROM tag t
        LEFT JOIN item_tag it ON t.id = it.tag_id
        LEFT JOIN item i ON it.item_id = i.id
        LEFT JOIN container c ON i.container_id = c.id
        WHERE 
            (c.user_id = $2 OR i.container_id IS NULL) AND
            to_tsvector('english', t.name) @@ plainto_tsquery('english', $1)
        GROUP BY t.id
        ORDER BY rank DESC;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing tag search: %v", err)
    }
    defer rows.Close()

    var results TagSearchResults
    for rows.Next() {
        var result TagSearchResult
        var itemsJSON []byte

        err := rows.Scan(
            &result.ID,
            &result.Name,
            &result.Colour,
            &result.CreatedAt,
            &result.UpdatedAt,
            &result.Rank,
            &itemsJSON,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning tag search result: %v", err)
        }

        if len(itemsJSON) > 0 {
            if err := json.Unmarshal(itemsJSON, &result.Items); err != nil {
                return nil, fmt.Errorf("error unmarshaling items: %v", err)
            }
        }

        results = append(results, result)
    }

    return results, nil
}