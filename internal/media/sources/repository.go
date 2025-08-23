package sources

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/chrisabs/cadence/internal/media/sources/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSource(source *entities.Source) error {
	query := `
		INSERT INTO media_source (id, name, color, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	now := time.Now().UTC()
	_, err := r.db.Exec(query, source.ID, source.Name, source.Color, source.Category, now, now)
	return err
}

func (r *Repository) GetSourceByID(sourceID models.SourceID) (*entities.Source, error) {
	query := `
		SELECT id, name, color, category, created_at, updated_at, is_deleted, deleted_at, deleted_by
		FROM media_source
		WHERE id = $1 AND is_deleted = false`
	
	var source entities.Source
	var deletedAt sql.NullTime
	var deletedBy sql.NullString
	
	err := r.db.QueryRow(query, sourceID).Scan(
		&source.ID, &source.Name, &source.Color, &source.Category,
		&source.CreatedAt, &source.UpdatedAt, &source.IsDeleted,
		&deletedAt, &deletedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("source not found")
		}
		return nil, err
	}
	
	if deletedAt.Valid {
		source.DeletedAt = &deletedAt.Time
	}
	if deletedBy.Valid {
		source.DeletedBy = (*models.ProfileID)(&deletedBy.String)
	}
	
	return &source, nil
}

func (r *Repository) UpdateSource(source *entities.Source) error {
	query := `
		UPDATE media_source 
		SET name = $2, color = $3, category = $4, updated_at = $5
		WHERE id = $1 AND is_deleted = false`
	
	now := time.Now().UTC()
	result, err := r.db.Exec(query, source.ID, source.Name, source.Color, source.Category, now)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("source not found")
	}
	
	return nil
}

func (r *Repository) DeleteSource(sourceID models.SourceID, deletedBy models.ProfileID) error {
	query := `
		UPDATE media_source 
		SET is_deleted = true, deleted_at = $2, deleted_by = $3
		WHERE id = $1 AND is_deleted = false`
	
	now := time.Now().UTC()
	result, err := r.db.Exec(query, sourceID, now, deletedBy)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("source not found")
	}
	
	return nil
}

func (r *Repository) GetAllSources(params entities.SourceSearchParams) (*entities.SourceSearchResponse, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	conditions = append(conditions, "is_deleted = false")
	
	if params.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, *params.Category)
		argIndex++
	}
	
	whereClause := strings.Join(conditions, " AND ")
	

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM media_source WHERE %s", whereClause)
	
	mainQuery := fmt.Sprintf(`
		SELECT id, name, color, category, created_at, updated_at, is_deleted, deleted_at, deleted_by
		FROM media_source 
		WHERE %s 
		ORDER BY category, name`, whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("error getting total count: %v", err)
	}
	
	if params.Limit != nil {
		mainQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, *params.Limit)
		argIndex++
		
		if params.Offset != nil {
			mainQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, *params.Offset)
		}
	}
	
	rows, err := r.db.Query(mainQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var sources []entities.Source
	for rows.Next() {
		var source entities.Source
		var deletedAt sql.NullTime
		var deletedBy sql.NullString
		
		err := rows.Scan(
			&source.ID, &source.Name, &source.Color, &source.Category,
			&source.CreatedAt, &source.UpdatedAt, &source.IsDeleted,
			&deletedAt, &deletedBy,
		)
		if err != nil {
			return nil, err
		}
		
		if deletedAt.Valid {
			source.DeletedAt = &deletedAt.Time
		}
		if deletedBy.Valid {
			source.DeletedBy = (*models.ProfileID)(&deletedBy.String)
		}
		
		sources = append(sources, source)
	}
	
	limit := 100 
	if params.Limit != nil {
		limit = *params.Limit
	}
	
	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}
	
	hasMore := (offset + limit) < total
	
	return &entities.SourceSearchResponse{
		Data:   sources,
		Total:  total,
		Limit:  limit,
		Offset: offset,
		HasMore: hasMore,
	}, nil
}

func (r *Repository) GetMaterialCountBySource(sourceID models.SourceID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM material
		WHERE $1 = ANY(source_ids) AND is_deleted = false`
	
	var count int
	err := r.db.QueryRow(query, sourceID).Scan(&count)
	return count, err
}
