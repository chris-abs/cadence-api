package search

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/chrisabs/storage/internal/models"
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
            CASE
                WHEN name ILIKE $1 THEN 1.0
                WHEN name ILIKE $1 || '%' THEN 0.8
                WHEN name ILIKE '%' || $1 || '%' THEN 0.6
                ELSE COALESCE(ts_rank(to_tsvector('english', name || ' ' || COALESCE(description, '')), 
                    websearch_to_tsquery('english', $1)), 0.0)
            END as rank,
            NULL as container_name,
            name as workspace_name,
            NULL as colour
        FROM workspace 
        WHERE 
            user_id = $2 AND
            (
                name ILIKE $1 OR
                to_tsvector('english', name || ' ' || COALESCE(description, '')) @@ 
                websearch_to_tsquery('english', $1)
            )
    ),
    container_matches AS (
        SELECT 
            'container' as type,
            c.id,
            c.name,
            COALESCE(c.location, '') as description,
            CASE
                WHEN c.name ILIKE $1 THEN 1.0
                WHEN c.name ILIKE $1 || '%' THEN 0.8
                WHEN c.name ILIKE '%' || $1 || '%' THEN 0.6
                ELSE COALESCE(ts_rank(to_tsvector('english', c.name), 
                    websearch_to_tsquery('english', $1)), 0.0)
            END as rank,
            NULL as container_name,
            w.name as workspace_name,
            NULL as colour
        FROM container c
        LEFT JOIN workspace w ON c.workspace_id = w.id
        WHERE 
            c.user_id = $2 AND
            (
                c.name ILIKE $1 OR
                to_tsvector('english', c.name) @@ websearch_to_tsquery('english', $1)
            )
    ),
    item_matches AS (
        SELECT 
            'item' as type,
            i.id,
            i.name,
            i.description,
            CASE
                WHEN i.name ILIKE $1 THEN 1.0
                WHEN i.name ILIKE $1 || '%' THEN 0.8
                WHEN i.name ILIKE '%' || $1 || '%' OR i.description ILIKE '%' || $1 || '%' THEN 0.6
                ELSE COALESCE(ts_rank(to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')), 
                    websearch_to_tsquery('english', $1)), 0.0)
            END as rank,
            c.name as container_name,
            NULL as workspace_name,
            NULL as colour
        FROM item i
        LEFT JOIN container c ON i.container_id = c.id
        WHERE 
            c.user_id = $2 AND
            (
                i.name ILIKE $1 OR
                i.description ILIKE '%' || $1 || '%' OR
                to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')) @@ 
                websearch_to_tsquery('english', $1)
            )
    ),
    tag_matches AS (
        SELECT DISTINCT
            'tag' as type,
            t.id,
            t.name,
            '' as description,
            CASE
                WHEN t.name ILIKE $1 THEN 1.0
                WHEN t.name ILIKE $1 || '%' THEN 0.8
                WHEN t.name ILIKE '%' || $1 || '%' THEN 0.6
                ELSE COALESCE(ts_rank(to_tsvector('english', t.name), 
                    websearch_to_tsquery('english', $1)), 0.0)
            END as rank,
            NULL as container_name,
            NULL as workspace_name,
            t.colour
        FROM tag t
        WHERE t.name ILIKE $1 OR t.name ILIKE $1 || '%' OR t.name ILIKE '%' || $1 || '%'
    ),
    tagged_items AS (
        SELECT DISTINCT
            'tagged_item' as type,
            i.id,
            i.name,
            i.description,
            CASE
                WHEN t.name ILIKE $1 THEN 0.9
                WHEN t.name ILIKE $1 || '%' THEN 0.7
                WHEN t.name ILIKE '%' || $1 || '%' THEN 0.5
                ELSE COALESCE(ts_rank(to_tsvector('english', t.name), 
                    websearch_to_tsquery('english', $1)), 0.0)
            END as rank,
            c.name as container_name,
            w.name as workspace_name,
            NULL as colour
        FROM item i
        INNER JOIN item_tag it ON i.id = it.item_id
        INNER JOIN tag t ON it.tag_id = t.id
        LEFT JOIN container c ON i.container_id = c.id
        LEFT JOIN workspace w ON c.workspace_id = w.id
        WHERE 
            (c.user_id = $2 OR i.container_id IS NULL) AND
            (
                t.name ILIKE $1 OR
                t.name ILIKE $1 || '%' OR
                t.name ILIKE '%' || $1 || '%' OR
                to_tsvector('english', t.name) @@ websearch_to_tsquery('english', $1)
            ) AND
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
            w.id, w.name, w.description, w.user_id, w.created_at, w.updated_at,
            CASE
                WHEN w.name ILIKE $1 THEN 1.0
                WHEN w.name ILIKE $1 || '%' THEN 0.8
                WHEN w.name ILIKE '%' || $1 || '%' OR w.description ILIKE '%' || $1 || '%' THEN 0.6
                ELSE COALESCE(ts_rank(to_tsvector('english', w.name || ' ' || COALESCE(w.description, '')), 
                    websearch_to_tsquery('english', $1)), 0.0)
            END as rank
        FROM workspace w
        WHERE 
            w.user_id = $2 AND
            (
                w.name ILIKE $1 OR
                w.description ILIKE '%' || $1 || '%' OR
                to_tsvector('english', w.name || ' ' || COALESCE(w.description, '')) @@ 
                websearch_to_tsquery('english', $1)
            )
        ORDER BY rank DESC;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing workspace search: %v", err)
    }
    defer rows.Close()

    var results WorkspaceSearchResults
    for rows.Next() {
        var result WorkspaceSearchResult
        err := rows.Scan(
            &result.ID,
            &result.Name,
            &result.Description,
            &result.UserID,
            &result.CreatedAt,
            &result.UpdatedAt,
            &result.Rank,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning workspace search result: %v", err)
        }

        results = append(results, result)
    }

    return results, nil
}

func (r *Repository) SearchContainers(query string, userID int) (ContainerSearchResults, error) {
    // First, do a quick check for matching containers
    quickCheckQuery := `
        SELECT EXISTS (
            SELECT 1
            FROM container c
            WHERE 
                c.user_id = $2 AND
                (
                    LOWER(c.name) = LOWER($1) OR
                    c.name ILIKE $1 || '%' OR
                    c.name ILIKE '%' || $1 || '%' OR
                    c.location ILIKE '%' || $1 || '%' OR
                    c.description ILIKE '%' || $1 || '%' OR
                    similarity(c.name, $1) > 0.3
                )
        );`

    var hasResults bool
    err := r.db.QueryRow(quickCheckQuery, query, userID).Scan(&hasResults)
    if err != nil {
        return nil, fmt.Errorf("error checking for results: %v", err)
    }

    if !hasResults {
        return ContainerSearchResults{}, nil
    }

    sqlQuery := `
        WITH ranked_containers AS (
            SELECT 
                c.id,
                c.name,
                c.description,
                c.qr_code,
                c.qr_code_image,
                c.number,
                c.location,
                c.user_id,
                c.workspace_id,
                c.created_at,
                c.updated_at,
                (
                    CASE
                        WHEN LOWER(c.name) = LOWER($1) THEN 100.0
                        WHEN c.name ~* ('\m' || $1 || '\M') THEN 90.0
                        WHEN c.name ILIKE $1 || '%' THEN 80.0
                        WHEN c.name ILIKE '%' || $1 || '%' THEN 60.0
                        WHEN c.description ILIKE '%' || $1 || '%' THEN 50.0
                        WHEN c.location ILIKE '%' || $1 || '%' THEN 40.0
                        WHEN similarity(c.name, $1) > 0.3 THEN similarity(c.name, $1) * 30.0
                        ELSE COALESCE(
                            ts_rank(
                                to_tsvector('english', 
                                    c.name || ' ' || 
                                    COALESCE(c.description, '') || ' ' || 
                                    COALESCE(c.location, '')
                                ),
                                websearch_to_tsquery('english', $1)
                            ) * 20.0,
                            0.0
                        )
                    END
                    +
                    CASE 
                        WHEN c.updated_at > NOW() - INTERVAL '7 days' THEN 5.0 
                        ELSE 0.0 
                    END
                ) as rank
            FROM container c
            WHERE 
                c.user_id = $2 AND
                (
                    LOWER(c.name) = LOWER($1) OR
                    c.name ~* ('\m' || $1 || '\M') OR
                    c.name ILIKE $1 || '%' OR
                    c.name ILIKE '%' || $1 || '%' OR
                    c.description ILIKE '%' || $1 || '%' OR
                    c.location ILIKE '%' || $1 || '%' OR
                    similarity(c.name, $1) > 0.3 OR
                    to_tsvector('english', 
                        c.name || ' ' || 
                        COALESCE(c.description, '') || ' ' || 
                        COALESCE(c.location, '')
                    ) @@ websearch_to_tsquery('english', $1)
                )
        )
        SELECT 
            rc.*,
            COALESCE(
                jsonb_build_object(
                    'id', w.id,
                    'name', w.name,
                    'description', w.description
                ),
                NULL
            ) as workspace
        FROM ranked_containers rc
        LEFT JOIN workspace w ON rc.workspace_id = w.id
        ORDER BY rc.rank DESC
        LIMIT 50;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing container search: %v", err)
    }
    defer rows.Close()

    var results ContainerSearchResults
    for rows.Next() {
        var result ContainerSearchResult
        var workspaceJSON []byte

        err := rows.Scan(
            &result.ID,
            &result.Name,
            &result.Description,
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
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning container search result: %v", err)
        }

        if len(workspaceJSON) > 0 {
            if err := json.Unmarshal(workspaceJSON, &result.Workspace); err != nil {
                return nil, fmt.Errorf("error unmarshaling workspace: %v", err)
            }
        }

        results = append(results, result)
    }

    return results, nil
}

func (r *Repository) SearchItems(query string, userID int) (ItemSearchResults, error) {
    quickCheckQuery := `
        SELECT EXISTS (
            SELECT 1
            FROM item i
            LEFT JOIN container c ON i.container_id = c.id
            WHERE 
                (c.user_id = $2 OR i.container_id IS NULL) AND
                (
                    LOWER(i.name) = LOWER($1) OR
                    i.name ~* ('\m' || $1 || '\M') OR
                    i.name ILIKE $1 || '%' OR
                    i.name ILIKE '%' || $1 || '%' OR
                    i.description ILIKE '%' || $1 || '%' OR
                    similarity(i.name, $1) > 0.3 OR
                    to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')) @@ 
                    websearch_to_tsquery('english', $1)
                )
        );`

    var hasResults bool
    err := r.db.QueryRow(quickCheckQuery, query, userID).Scan(&hasResults)
    if err != nil {
        return nil, fmt.Errorf("error checking for results: %v", err)
    }

    if !hasResults {
        return ItemSearchResults{}, nil
    }

    sqlQuery := `
        WITH ranked_items AS (
            SELECT 
                i.id,
                i.name,
                i.description,
                i.quantity,
                i.container_id,
                i.created_at,
                i.updated_at,
                (
                    CASE
                        WHEN LOWER(i.name) = LOWER($1) THEN 100.0
                        WHEN i.name ~* ('\m' || $1 || '\M') THEN 90.0
                        WHEN i.name ILIKE $1 || '%' THEN 80.0
                        WHEN i.name ILIKE '%' || $1 || '%' THEN 60.0
                        WHEN i.description ILIKE '%' || $1 || '%' THEN 40.0
                        WHEN similarity(i.name, $1) > 0.3 THEN similarity(i.name, $1) * 30.0
                        ELSE COALESCE(
                            ts_rank(
                                to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')),
                                websearch_to_tsquery('english', $1)
                            ) * 20.0,
                            0.0
                        )
                    END
                    +
                    CASE 
                        WHEN i.updated_at > NOW() - INTERVAL '7 days' THEN 5.0 
                        ELSE 0.0 
                    END
                ) as rank
            FROM item i
            LEFT JOIN container c ON i.container_id = c.id
            WHERE 
                (c.user_id = $2 OR i.container_id IS NULL) AND
                (
                    LOWER(i.name) = LOWER($1) OR
                    i.name ~* ('\m' || $1 || '\M') OR
                    i.name ILIKE $1 || '%' OR
                    i.name ILIKE '%' || $1 || '%' OR
                    i.description ILIKE '%' || $1 || '%' OR
                    similarity(i.name, $1) > 0.3 OR
                    to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')) @@ 
                    websearch_to_tsquery('english', $1)
                )
        )`

    sqlQuery += `
        , item_images AS (
            SELECT 
                img.item_id,
                jsonb_agg(
                    jsonb_build_object(
                        'id', img.id,
                        'url', img.url,
                        'display_order', img.display_order
                    ) ORDER BY img.display_order
                ) as images
            FROM item_image img
            JOIN ranked_items ri ON img.item_id = ri.id
            GROUP BY img.item_id
        )
        SELECT 
            i.id,
            i.name,
            i.description,
            i.quantity,
            i.container_id,
            i.created_at,
            i.updated_at,
            i.rank,
            jsonb_build_object(
                'id', c.id,
                'name', c.name,
                'location', c.location,
                'workspace_id', c.workspace_id
            ) as container,
            COALESCE(
                jsonb_agg(
                    DISTINCT jsonb_build_object(
                        'id', t.id,
                        'name', t.name,
                        'colour', t.colour
                    )
                ) FILTER (WHERE t.id IS NOT NULL),
                '[]'
            ) as tags,
            COALESCE(ii.images, '[]'::jsonb) as images
        FROM ranked_items i
        LEFT JOIN container c ON i.container_id = c.id
        LEFT JOIN item_tag it ON i.id = it.item_id
        LEFT JOIN tag t ON it.tag_id = t.id
        LEFT JOIN item_images ii ON i.id = ii.item_id
        GROUP BY 
            i.id, i.name, i.description, i.quantity, i.container_id, 
            i.created_at, i.updated_at, i.rank,
            c.id, c.name, c.location, c.workspace_id,
            ii.images
        ORDER BY i.rank DESC
        LIMIT 50;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing item search: %v", err)
    }
    defer rows.Close()

    var results ItemSearchResults
    for rows.Next() {
        var result ItemSearchResult
        var containerJSON, tagsJSON, imagesJSON []byte

        err := rows.Scan(
            &result.ID,
            &result.Name,
            &result.Description,
            &result.Quantity,
            &result.ContainerID,
            &result.CreatedAt,
            &result.UpdatedAt,
            &result.Rank,
            &containerJSON,
            &tagsJSON,
            &imagesJSON,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning item search result: %v", err)
        }

        if len(containerJSON) > 0 {
            if err := json.Unmarshal(containerJSON, &result.Container); err != nil {
                return nil, fmt.Errorf("error unmarshaling container: %v", err)
            }
        }

        if err := json.Unmarshal(tagsJSON, &result.Tags); err != nil {
            return nil, fmt.Errorf("error unmarshaling tags: %v", err)
        }

        if err := json.Unmarshal(imagesJSON, &result.Images); err != nil {
            return nil, fmt.Errorf("error unmarshaling images: %v", err)
        }

        results = append(results, result)
    }

    return results, nil
}

func (r *Repository) SearchTags(query string, userID int) (TagSearchResults, error) {
    quickCheckQuery := `
        SELECT EXISTS (
            SELECT 1
            FROM tag t
            LEFT JOIN item_tag it ON t.id = it.tag_id
            LEFT JOIN item i ON it.item_id = i.id
            LEFT JOIN container c ON i.container_id = c.id
            WHERE 
                (c.user_id = $2 OR i.container_id IS NULL) AND
                (
                    t.name ILIKE $1 OR
                    t.name ILIKE $1 || '%' OR
                    t.name ILIKE '%' || $1 || '%' OR
                    similarity(t.name, $1) > 0.3
                )
        );`

    var hasResults bool
    err := r.db.QueryRow(quickCheckQuery, query, userID).Scan(&hasResults)
    if err != nil {
        return nil, fmt.Errorf("error checking for results: %v", err)
    }

    if !hasResults {
        return TagSearchResults{}, nil
    }

    sqlQuery := `
        WITH ranked_tags AS (
            SELECT DISTINCT ON (t.id)
                t.id,
                t.name,
                t.colour,
                t.description,
                t.created_at,
                t.updated_at,
                (
                    CASE
                        WHEN LOWER(t.name) = LOWER($1) THEN 100.0
                        WHEN t.name ILIKE $1 || '%' THEN 80.0
                        WHEN t.name ILIKE '%' || $1 || '%' THEN 60.0
                        WHEN similarity(t.name, $1) > 0.3 THEN similarity(t.name, $1) * 30.0
                        ELSE 0.0
                    END
                    +
                    CASE 
                        WHEN t.updated_at > NOW() - INTERVAL '7 days' THEN 5.0 
                        ELSE 0.0 
                    END
                ) as rank
            FROM tag t
            LEFT JOIN item_tag it ON t.id = it.tag_id
            LEFT JOIN item i ON it.item_id = i.id
            LEFT JOIN container c ON i.container_id = c.id
            WHERE 
                (c.user_id = $2 OR i.container_id IS NULL) AND
                (
                    t.name ILIKE $1 OR
                    t.name ILIKE $1 || '%' OR
                    t.name ILIKE '%' || $1 || '%' OR
                    similarity(t.name, $1) > 0.3
                )
        )
        SELECT 
            rt.id,
            rt.name,
            rt.colour,
            rt.description,
            rt.created_at,
            rt.updated_at,
            rt.rank,
            COALESCE(
                jsonb_agg(
                    CASE 
                        WHEN i.id IS NOT NULL THEN
                            jsonb_build_object(
                                'id', i.id,
                                'name', i.name,
                                'quantity', i.quantity,
                                'container_id', i.container_id
                            )
                        ELSE NULL 
                    END
                ) FILTER (WHERE i.id IS NOT NULL),
                '[]'::jsonb
            ) as items
        FROM ranked_tags rt
        LEFT JOIN item_tag it ON rt.id = it.tag_id
        LEFT JOIN item i ON it.item_id = i.id
        GROUP BY rt.id, rt.name, rt.colour, rt.description, rt.created_at, rt.updated_at, rt.rank
        ORDER BY rt.rank DESC
        LIMIT 50;`

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
            &result.Description,
            &result.CreatedAt,
            &result.UpdatedAt,
            &result.Rank,
            &itemsJSON,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning tag search result: %v", err)
        }

        if err := json.Unmarshal(itemsJSON, &result.Items); err != nil {
            return nil, fmt.Errorf("error unmarshaling items: %v", err)
        }

        results = append(results, result)
    }

    return results, nil
}

func (r *Repository) FindContainerByQR(qrCode string, userID int) (*models.Container, error) {
   query := `
       SELECT 
           c.*,
           jsonb_build_object(
               'id', w.id,
               'name', w.name,
               'description', w.description,
               'userId', w.user_id,
               'createdAt', w.created_at,
               'updatedAt', w.updated_at
           ) as workspace
       FROM container c
       LEFT JOIN workspace w ON c.workspace_id = w.id
       WHERE c.qr_code = $1 AND c.user_id = $2
       LIMIT 1`

   container := new(models.Container)
   var workspaceJSON []byte
   
   err := r.db.QueryRow(query, qrCode, userID).Scan(
       &container.ID,
       &container.Name,
       &container.QRCode,
       &container.QRCodeImage,
       &container.Number,
       &container.Location,
       &container.UserID,
       &container.WorkspaceID,
       &container.CreatedAt,
       &container.UpdatedAt,
       &workspaceJSON,
   )

   if err == sql.ErrNoRows {
       return nil, fmt.Errorf("container not found")
   }
   if err != nil {
       return nil, fmt.Errorf("error finding container: %v", err)
   }

   if len(workspaceJSON) > 0 {
       if err := json.Unmarshal(workspaceJSON, &container.Workspace); err != nil {
           return nil, fmt.Errorf("error unmarshaling workspace: %v", err)
       }
   }

   return container, nil
}
       