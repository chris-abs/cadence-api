package search

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
        name as workspace_name
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
        w.name as workspace_name
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
        w.name as workspace_name
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
        NULL as workspace_name
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
        w.name as workspace_name
    FROM item i
    INNER JOIN item_tag it ON i.id = it.item_id
    INNER JOIN tag t ON it.tag_id = t.id
    LEFT JOIN container c ON i.container_id = c.id
    LEFT JOIN workspace w ON c.workspace_id = w.id
    WHERE 
        (c.user_id = $2 OR i.container_id IS NULL) AND
        to_tsvector('english', t.name) @@ plainto_tsquery('english', $1) AND
        i.id NOT IN (SELECT id FROM item_matches) -- Exclude direct matches
)
SELECT type, id, name, description, rank, container_name, workspace_name FROM (
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
        var containerName, workspaceName sql.NullString
        err := rows.Scan(
            &result.Type,
            &result.ID,
            &result.Name,
            &result.Description,
            &result.Rank,
            &containerName,
            &workspaceName,
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
            'workspace' as type,
            id,
            name,
            description,
            ts_rank(to_tsvector('english', name || ' ' || COALESCE(description, '')), 
                   plainto_tsquery('english', $1)) as rank
        FROM workspace 
        WHERE 
            user_id = $2 AND
            to_tsvector('english', name || ' ' || COALESCE(description, '')) @@ 
            plainto_tsquery('english', $1)
        ORDER BY rank DESC;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing workspace search: %v", err)
    }
    defer rows.Close()

    var results WorkspaceSearchResults
    for rows.Next() {
        var result SearchResult
        err := rows.Scan(
            &result.Type,
            &result.ID,
            &result.Name,
            &result.Description,
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
    sqlQuery := `
        SELECT 
            'container' as type,
            c.id,
            c.name,
            COALESCE(c.description, '') as description,
            ts_rank(to_tsvector('english', c.name || ' ' || COALESCE(c.description, '')), 
                   plainto_tsquery('english', $1)) as rank,
            w.name as workspace_name
        FROM container c
        LEFT JOIN workspace w ON c.workspace_id = w.id 
        WHERE 
            c.user_id = $2 AND
            to_tsvector('english', c.name || ' ' || COALESCE(c.description, '')) @@ 
            plainto_tsquery('english', $1)
        ORDER BY rank DESC;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing container search: %v", err)
    }
    defer rows.Close()

    var results ContainerSearchResults
    for rows.Next() {
        var result SearchResult
        var workspaceName sql.NullString
        err := rows.Scan(
            &result.Type,
            &result.ID,
            &result.Name,
            &result.Description,
            &result.Rank,
            &workspaceName,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning container search result: %v", err)
        }
        if workspaceName.Valid {
            result.WorkspaceName = &workspaceName.String
        }
        results = append(results, result)
    }

    return results, nil
}

func (r *Repository) SearchItems(query string, userID int) (ItemSearchResults, error) {
    sqlQuery := `
        SELECT 
            'item' as type,
            i.id,
            i.name,
            i.description,
            ts_rank(to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')), 
                   plainto_tsquery('english', $1)) as rank,
            c.name as container_name
        FROM item i
        LEFT JOIN container c ON i.container_id = c.id
        WHERE 
            (c.user_id = $2 OR i.container_id IS NULL) AND
            to_tsvector('english', i.name || ' ' || COALESCE(i.description, '')) @@ 
            plainto_tsquery('english', $1)
        ORDER BY rank DESC;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing item search: %v", err)
    }
    defer rows.Close()

    var results ItemSearchResults
    for rows.Next() {
        var result SearchResult
        var containerName sql.NullString
        err := rows.Scan(
            &result.Type,
            &result.ID,
            &result.Name,
            &result.Description,
            &result.Rank,
            &containerName,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning item search result: %v", err)
        }
        if containerName.Valid {
            result.ContainerName = &containerName.String
        }
        results = append(results, result)
    }

    return results, nil
}

func (r *Repository) SearchTags(query string, userID int) (TagSearchResults, error) {
    sqlQuery := `
        SELECT DISTINCT
            'tag' as type,
            t.id,
            t.name,
            '' as description,
            ts_rank(to_tsvector('english', t.name), plainto_tsquery('english', $1)) as rank
        FROM tag t
        LEFT JOIN item_tag it ON t.id = it.tag_id
        LEFT JOIN item i ON it.item_id = i.id
        LEFT JOIN container c ON i.container_id = c.id
        WHERE 
            (c.user_id = $2 OR i.container_id IS NULL) AND
            to_tsvector('english', t.name) @@ plainto_tsquery('english', $1)
        ORDER BY rank DESC;`

    rows, err := r.db.Query(sqlQuery, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error executing tag search: %v", err)
    }
    defer rows.Close()

    var results TagSearchResults
    for rows.Next() {
        var result SearchResult
        err := rows.Scan(
            &result.Type,
            &result.ID,
            &result.Name,
            &result.Description,
            &result.Rank,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning tag search result: %v", err)
        }
        results = append(results, result)
    }

    return results, nil
}