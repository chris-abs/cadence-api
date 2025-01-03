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
        WITH container_matches AS (
            SELECT 
                'container' as type,
                id,
                name,
                '' as description,
                ts_rank(to_tsvector('english', name), plainto_tsquery('english', $1)) as rank
            FROM container 
            WHERE 
                user_id = $2 AND
                to_tsvector('english', name) @@ plainto_tsquery('english', $1)
        ),
        item_matches AS (
            SELECT 
                'item' as type,
                i.id,
                i.name,
                i.description,
                ts_rank(to_tsvector('english', i.name || ' ' || i.description), 
                       plainto_tsquery('english', $1)) as rank
            FROM item i
            JOIN container c ON i.container_id = c.id
            WHERE 
                c.user_id = $2 AND
                to_tsvector('english', i.name || ' ' || i.description) @@ 
                plainto_tsquery('english', $1)
        ),
        tag_matches AS (
            SELECT DISTINCT
                'tag' as type,
                t.id,
                t.name,
                '' as description,
                ts_rank(to_tsvector('english', t.name), plainto_tsquery('english', $1)) as rank
            FROM tag t
            JOIN item_tag it ON t.id = it.tag_id
            JOIN item i ON it.item_id = i.id
            JOIN container c ON i.container_id = c.id
            WHERE 
                c.user_id = $2 AND
                to_tsvector('english', t.name) @@ plainto_tsquery('english', $1)
        )
        SELECT * FROM (
            SELECT * FROM container_matches
            UNION ALL
            SELECT * FROM item_matches
            UNION ALL
            SELECT * FROM tag_matches
        ) combined_results
        ORDER BY rank DESC;`

	rows, err := r.db.Query(sqlQuery, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %v", err)
	}
	defer rows.Close()

	response := &SearchResponse{
		Containers: make([]SearchResult, 0),
		Items:      make([]SearchResult, 0),
		Tags:       make([]SearchResult, 0),
	}

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
			return nil, fmt.Errorf("error scanning search result: %v", err)
		}

		switch result.Type {
		case "container":
			response.Containers = append(response.Containers, result)
		case "item":
			response.Items = append(response.Items, result)
		case "tag":
			response.Tags = append(response.Tags, result)
		}
	}

	return response, nil
}
