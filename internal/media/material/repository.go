package material

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chrisabs/cadence/internal/media/material/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(profileID models.ProfileID, familyID models.FamilyID, req *CreateMaterialRequest) (*entities.Material, error) {
	materialID := models.NewMaterialID()
	
	query := `
		INSERT INTO material (
			id, name, type, genre, release_year, runtime, poster_url, source_ids,
			watch_with, status, priority, notes, profile_id, family_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at`

	sourceIDsJSON, err := json.Marshal(req.SourceIDs)
	if err != nil {
		return nil, fmt.Errorf("error marshaling source IDs: %v", err)
	}

	now := time.Now().UTC()
	var createdAt, updatedAt time.Time

	err = r.db.QueryRow(
		query,
		materialID, req.Name, req.Type, req.Genre, req.ReleaseYear, req.Runtime, req.PosterURL,
		sourceIDsJSON, req.WatchWith, req.Status, req.Priority, req.Notes,
		profileID, familyID, now, now,
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("error creating media: %v", err)
	}

	return r.GetByID(materialID, familyID)
}

func (r *Repository) GetByID(id models.MaterialID, familyID models.FamilyID) (*entities.Material, error) {
	query := `
		SELECT m.id, m.name, m.type, m.genre, m.release_year, m.runtime, m.poster_url,
			   m.source_ids, m.watch_with, m.status, m.priority, m.notes,
			   m.profile_id, m.family_id, m.created_at, m.updated_at
		FROM media m
		WHERE m.id = $1 AND m.family_id = $2 AND m.is_deleted = false`

	material := new(entities.Material)
	var sourceIDsJSON []byte

	err := r.db.QueryRow(query, id, familyID).Scan(
		&material.ID, &material.Name, &material.Type, &material.Genre, &material.ReleaseYear,
		&material.Runtime, &material.PosterURL, &sourceIDsJSON, &material.WatchWith,
		&material.Status, &material.Priority, &material.Notes, &material.ProfileID,
		&material.FamilyID, &material.CreatedAt, &material.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("material not found")
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(sourceIDsJSON, &material.SourceIDs); err != nil {
		return nil, fmt.Errorf("error parsing source IDs: %v", err)
	}


	return material, nil
}

func (r *Repository) Search(familyID models.FamilyID, req *MaterialSearchRequest) ([]entities.Material, int, error) {
	profileID := req.ProfileID
	if profileID == nil {
		return nil, 0, fmt.Errorf("profile ID is required")
	}

	conditions := []string{"m.family_id = $1", "m.profile_id = $2", "m.is_deleted = false"}
	args := []interface{}{familyID, *profileID}
	argIndex := 3

	if req.Query != "" {
		conditions = append(conditions, fmt.Sprintf(`(
			LOWER(m.name) = LOWER($%d) OR
			m.name ILIKE $%d || '%%' OR
			m.name ILIKE '%%' || $%d || '%%' OR
			m.notes ILIKE '%%' || $%d || '%%' OR
			to_tsvector('english', m.name || ' ' || COALESCE(m.notes, '')) @@ 
			websearch_to_tsquery('english', $%d)
		)`, argIndex, argIndex, argIndex, argIndex, argIndex))
		args = append(args, req.Query)
		argIndex++
	}

	if req.Type != "" {
		conditions = append(conditions, fmt.Sprintf("m.type = $%d", argIndex))
		args = append(args, req.Type)
		argIndex++
	}

	if req.Genre != "" {
		conditions = append(conditions, fmt.Sprintf("m.genre = $%d", argIndex))
		args = append(args, req.Genre)
		argIndex++
	}

	if req.SourceID != "" {
		conditions = append(conditions, fmt.Sprintf("m.source_ids ? $%d", argIndex))
		args = append(args, string(req.SourceID))
		argIndex++
	}

	if req.WatchWith != "" {
		conditions = append(conditions, fmt.Sprintf("m.watch_with = $%d", argIndex))
		args = append(args, req.WatchWith)
		argIndex++
	}

	if req.Status != "" {
		conditions = append(conditions, fmt.Sprintf("m.status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++
	}

	if req.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("m.priority = $%d", argIndex))
		args = append(args, req.Priority)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM material m WHERE %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []entities.Material{}, 0, nil
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		WITH ranked_media AS (
			SELECT 
				m.id, m.name, m.type, m.genre, m.release_year, m.runtime, m.poster_url,
				m.source_ids, m.watch_with, m.status, m.priority, m.notes,
				m.profile_id, m.family_id, m.created_at, m.updated_at,
				(
					CASE
						WHEN m.status = 'to_watch' THEN 3
						WHEN m.status = 'in_progress' THEN 4
						WHEN m.status = 'watching' THEN 5
						WHEN m.status = 'awaiting_release' THEN 2
						WHEN m.status = 'watched' THEN 1
						ELSE 0
					END +
					CASE
						WHEN m.priority = 'high' THEN 2
						WHEN m.priority = 'medium' THEN 1
						ELSE 0
					END
				) as rank
			FROM material m
			WHERE %s
		)
		SELECT * FROM ranked_media
		ORDER BY rank DESC, name
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing media search: %v", err)
	}
	defer rows.Close()

	var materialList []entities.Material
	for rows.Next() {
		var material entities.Material
		var sourceIDsJSON []byte
		var rank int

		err := rows.Scan(
			&material.ID, &material.Name, &material.Type, &material.Genre, &material.ReleaseYear,
			&material.Runtime, &material.PosterURL, &sourceIDsJSON, &material.WatchWith,
			&material.Status, &material.Priority, &material.Notes, &material.ProfileID,
			&material.FamilyID, &material.CreatedAt, &material.UpdatedAt, &rank,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning media result: %v", err)
		}

		if err := json.Unmarshal(sourceIDsJSON, &material.SourceIDs); err != nil {
			return nil, 0, fmt.Errorf("error parsing source IDs: %v", err)
		}


		materialList = append(materialList, material)
	}

	return materialList, total, nil
}

func (r *Repository) Update(id models.MaterialID, familyID models.FamilyID, profileID models.ProfileID, req *UpdateMaterialRequest) (*entities.Material, error) {
	sourceIDsJSON, err := json.Marshal(req.SourceIDs)
	if err != nil {
		return nil, fmt.Errorf("error marshaling source IDs: %v", err)
	}

	query := `
		UPDATE material
		SET name = $3, type = $4, genre = $5, release_year = $6, runtime = $7, 
			poster_url = $8, source_ids = $9, watch_with = $10, status = $11, 
			priority = $12, notes = $13, updated_at = $14
		WHERE id = $1 AND family_id = $2 AND profile_id = $15 AND is_deleted = false`

	result, err := r.db.Exec(
		query, id, familyID, req.Name, req.Type, req.Genre, req.ReleaseYear,
		req.Runtime, req.PosterURL, sourceIDsJSON, req.WatchWith, req.Status,
		req.Priority, req.Notes, time.Now().UTC(), profileID,
	)

	if err != nil {
		return nil, fmt.Errorf("error updating media: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("media not found or not owned by profile")
	}

	return r.GetByID(id, familyID)
}

func (r *Repository) UpdateStatus(id models.MaterialID, familyID models.FamilyID, profileID models.ProfileID, status entities.Status) error {
	query := `
		UPDATE material
		SET status = $3, updated_at = $4
		WHERE id = $1 AND family_id = $2 AND profile_id = $5 AND is_deleted = false`

	result, err := r.db.Exec(query, id, familyID, status, time.Now().UTC(), profileID)
	if err != nil {
		return fmt.Errorf("error updating media status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("media not found or not owned by profile")
	}

	return nil
}

func (r *Repository) Delete(id models.MaterialID, familyID models.FamilyID, profileID models.ProfileID, deletedBy models.ProfileID) error {
	query := `
		UPDATE material
		SET is_deleted = true, deleted_at = $4, deleted_by = $5, updated_at = $4
		WHERE id = $1 AND family_id = $2 AND profile_id = $3 AND is_deleted = false`

	result, err := r.db.Exec(query, id, familyID, profileID, time.Now().UTC(), deletedBy)
	if err != nil {
		return fmt.Errorf("error deleting media: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("media not found or not owned by profile")
	}

	return nil
}

func (r *Repository) GetStatusSummary(familyID models.FamilyID, profileID models.ProfileID) (*MaterialStatusSummaryResponse, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN status = 'to_watch' THEN 1 END) as to_watch,
			COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress,
			COUNT(CASE WHEN status = 'watching' THEN 1 END) as watching,
			COUNT(CASE WHEN status = 'awaiting_release' THEN 1 END) as awaiting_release,
			COUNT(CASE WHEN status = 'watched' THEN 1 END) as watched,
			COUNT(*) as total
		FROM material
		WHERE family_id = $1 AND profile_id = $2 AND is_deleted = false`

	summary := new(MaterialStatusSummaryResponse)
	err := r.db.QueryRow(query, familyID, profileID).Scan(
		&summary.ToWatch, &summary.InProgress, &summary.Watching,
		&summary.AwaitingRelease, &summary.Watched, &summary.Total,
	)

	if err != nil {
		return nil, fmt.Errorf("error getting status summary: %v", err)
	}

	return summary, nil
}
