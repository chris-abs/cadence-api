package material

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chrisabs/cadence/internal/media/material/entities"
	"github.com/chrisabs/cadence/internal/media/sources"
	"github.com/chrisabs/cadence/internal/models"
)

type Repository struct {
	db *sql.DB
	sourceRepo *sources.Repository
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
		sourceRepo: sources.NewRepository(db),
	}
}

func (r *Repository) Create(profileID models.ProfileID, familyID models.FamilyID, req *CreateMaterialRequest) (*entities.Material, error) {
	materialID := models.NewMaterialID()
	
	query := `
		INSERT INTO material (
			id, name, type, runtime, poster_url, source_ids, classification_id,
			watch_with, status, priority, notes, review_score,
			profile_id, family_id, created_at, updated_at,
			is_deleted, deleted_at, deleted_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
		RETURNING created_at, updated_at`

	sourceIDsJSON, err := json.Marshal(req.SourceIDs)
	if err != nil {
		return nil, fmt.Errorf("error marshaling source IDs: %v", err)
	}

	now := time.Now().UTC()
	var createdAt, updatedAt time.Time

	err = r.db.QueryRow(
		query,
		materialID, req.Name, req.Type, req.Runtime, req.PosterURL,
		sourceIDsJSON, req.ClassificationID, req.WatchWith, req.Status, req.Priority, req.Notes,
		nil, 
		profileID, familyID, now, now,
		false, nil, nil, 
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("error creating media: %v", err)
	}

	return r.GetByID(materialID, profileID)
}

func (r *Repository) GetByID(id models.MaterialID, profileID models.ProfileID) (*entities.Material, error) {
	query := `
		SELECT m.id, m.name, m.type, m.runtime, m.poster_url,
			   m.source_ids, m.classification_id, m.watch_with, m.status, m.priority, m.notes, m.review_score,
			   m.profile_id, m.family_id, m.created_at, m.updated_at
		FROM material m
		WHERE m.id = $1 AND m.profile_id = $2 AND m.is_deleted = false`
	
	material := new(entities.Material)
	var sourceIDsJSON []byte
	
	err := r.db.QueryRow(query, id, profileID).Scan(
		&material.ID, &material.Name, &material.Type,
		&material.Runtime, &material.PosterURL, &sourceIDsJSON, &material.ClassificationID, &material.WatchWith,
		&material.Status, &material.Priority, &material.Notes, &material.ReviewScore, &material.ProfileID,
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

	if len(material.SourceIDs) > 0 {
		sources, err := r.sourceRepo.GetSourcesByIDs(material.SourceIDs, material.ProfileID)
		if err != nil {
			return nil, fmt.Errorf("error fetching sources: %v", err)
		}
		material.Sources = sources
	}

	return material, nil
}

func (r *Repository) Search(familyID models.FamilyID, req *MaterialSearchRequest) (*MaterialSearchResponse, error) {
	profileID := req.ProfileID
	if profileID == nil {
		return nil, fmt.Errorf("profile ID is required")
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

	if req.Runtime != "" {
		conditions = append(conditions, fmt.Sprintf("m.runtime = $%d", argIndex))
		args = append(args, req.Runtime)
		argIndex++
	}

	if req.SourceID != "" {
		conditions = append(conditions, fmt.Sprintf("m.source_ids ? $%d", argIndex))
		args = append(args, string(req.SourceID))
		argIndex++
	}

	if req.ClassificationID != nil {
		conditions = append(conditions, fmt.Sprintf("m.classification_id = $%d", argIndex))
		args = append(args, *req.ClassificationID)
		argIndex++
	} else if !req.IncludeClassified {
		conditions = append(conditions, "m.classification_id IS NULL")
	}

	if req.WatchWith != "" {
		conditions = append(conditions, fmt.Sprintf("m.watch_with = $%d", argIndex))
		args = append(args, req.WatchWith)
		argIndex++
	}

	if req.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("m.priority = $%d", argIndex))
		args = append(args, req.Priority)
		argIndex++
	}

	if req.Status != "" {
		conditions = append(conditions, fmt.Sprintf("m.status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	orderBy := r.buildOrderByClause(req.SortBy)

	query := fmt.Sprintf(`
		SELECT 
			m.id, m.name, m.type, m.runtime, m.poster_url,
			m.source_ids, m.classification_id, m.watch_with, m.status, m.priority, m.notes, m.review_score,
			m.profile_id, m.family_id, m.created_at, m.updated_at
		FROM material m
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing media search: %v", err)
	}
	defer rows.Close()

	var materialList []entities.Material
	for rows.Next() {
		var material entities.Material
		var sourceIDsJSON []byte

		err := rows.Scan(
			&material.ID, &material.Name, &material.Type,
			&material.Runtime, &material.PosterURL, &sourceIDsJSON, &material.ClassificationID, &material.WatchWith,
			&material.Status, &material.Priority, &material.Notes, &material.ReviewScore, &material.ProfileID,
			&material.FamilyID, &material.CreatedAt, &material.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning media result: %v", err)
		}

		if err := json.Unmarshal(sourceIDsJSON, &material.SourceIDs); err != nil {
			return nil, fmt.Errorf("error parsing source IDs: %v", err)
		}

		if len(material.SourceIDs) > 0 {
			sources, err := r.sourceRepo.GetSourcesByIDs(material.SourceIDs, material.ProfileID)
			if err != nil {
				return nil, fmt.Errorf("error fetching sources: %v", err)
			}
			material.Sources = sources
		}

		materialList = append(materialList, material)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM material m WHERE %s", whereClause)
	var total int
	err = r.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("error getting total count: %v", err)
	}

	hasMore := offset+len(materialList) < total

	response := &MaterialSearchResponse{
		Material: materialList,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
		HasMore:  hasMore,
		SortBy:   req.SortBy,
	}

	return response, nil
}

func (r *Repository) buildOrderByClause(sortBy string) string {
	statusOrder := `ORDER BY 
		CASE 
			WHEN m.status = 'to_watch' THEN 1
			WHEN m.status = 'in_progress' THEN 2
			WHEN m.status = 'watching' THEN 3
			WHEN m.status = 'awaiting_release' THEN 4
			WHEN m.status = 'watched' THEN 5
		END,`

	var userSort string
	switch sortBy {
	case "alphabetical_asc":
		userSort = "m.name ASC"
	case "alphabetical_desc":
		userSort = "m.name DESC"
	case "recently_added":
		userSort = "m.created_at DESC"
	case "priority":
		userSort = `CASE 
			WHEN m.priority = 'high' THEN 1
			WHEN m.priority = 'medium' THEN 2
			WHEN m.priority = 'low' THEN 3
			ELSE 4
		END`
	case "highest_rating":
		userSort = "m.review_score DESC NULLS LAST"
	default:
		userSort = "m.review_score DESC NULLS LAST"
	}

	return statusOrder + " " + userSort + ", m.name ASC"
}

func (r *Repository) SearchAllColumns(familyID models.FamilyID, req *MaterialSearchRequest) (*MaterialSearchResponse, error) {
	profileID := req.ProfileID
	if profileID == nil {
		return nil, fmt.Errorf("profile ID is required")
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

	if req.Runtime != "" {
		conditions = append(conditions, fmt.Sprintf("m.runtime = $%d", argIndex))
		args = append(args, req.Runtime)
		argIndex++
	}

	if req.SourceID != "" {
		conditions = append(conditions, fmt.Sprintf("m.source_ids ? $%d", argIndex))
		args = append(args, string(req.SourceID))
		argIndex++
	}

	if req.ClassificationID != nil {
		conditions = append(conditions, fmt.Sprintf("m.classification_id = $%d", argIndex))
		args = append(args, *req.ClassificationID)
		argIndex++
	} else if !req.IncludeClassified {
		conditions = append(conditions, "m.classification_id IS NULL")
	}

	if req.WatchWith != "" {
		conditions = append(conditions, fmt.Sprintf("m.watch_with = $%d", argIndex))
		args = append(args, req.WatchWith)
		argIndex++
	}

	if req.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("m.priority = $%d", argIndex))
		args = append(args, req.Priority)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	orderBy := r.buildOrderByClause(req.SortBy)

	sqlLimit := limit * 5
	query := fmt.Sprintf(`
		SELECT 
			m.id, m.name, m.type, m.runtime, m.poster_url,
			m.source_ids, m.classification_id, m.watch_with, m.status, m.priority, m.notes, m.review_score,
			m.profile_id, m.family_id, m.created_at, m.updated_at
		FROM material m
		WHERE %s
		%s
		LIMIT $%d`, whereClause, orderBy, argIndex)

	args = append(args, sqlLimit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing media search: %v", err)
	}
	defer rows.Close()

	response := &MaterialSearchResponse{
		Total:    0,
		Limit:    limit,
		Offset:   0,
		HasMore:  false,
		SortBy:   req.SortBy,
	}

	totalCount := 0
	for rows.Next() {
		var material entities.Material
		var sourceIDsJSON []byte

		err := rows.Scan(
			&material.ID, &material.Name, &material.Type,
			&material.Runtime, &material.PosterURL, &sourceIDsJSON, &material.ClassificationID, &material.WatchWith,
			&material.Status, &material.Priority, &material.Notes, &material.ReviewScore, &material.ProfileID,
			&material.FamilyID, &material.CreatedAt, &material.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning media result: %v", err)
		}

		if err := json.Unmarshal(sourceIDsJSON, &material.SourceIDs); err != nil {
			return nil, fmt.Errorf("error parsing source IDs: %v", err)
		}

		if len(material.SourceIDs) > 0 {
			sources, err := r.sourceRepo.GetSourcesByIDs(material.SourceIDs, material.ProfileID)
			if err != nil {
				return nil, fmt.Errorf("error fetching sources: %v", err)
			}
			material.Sources = sources
		}

		switch material.Status {
		case entities.StatusToWatch:
			if len(response.Columns.ToWatch) < limit {
				response.Columns.ToWatch = append(response.Columns.ToWatch, material)
			}
		case entities.StatusInProgress:
			if len(response.Columns.InProgress) < limit {
				response.Columns.InProgress = append(response.Columns.InProgress, material)
			}
		case entities.StatusWatching:
			if len(response.Columns.Watching) < limit {
				response.Columns.Watching = append(response.Columns.Watching, material)
			}
		case entities.StatusAwaitingRelease:
			if len(response.Columns.AwaitingRelease) < limit {
				response.Columns.AwaitingRelease = append(response.Columns.AwaitingRelease, material)
			}
		case entities.StatusWatched:
			if len(response.Columns.Watched) < limit {
				response.Columns.Watched = append(response.Columns.Watched, material)
			}
		}

		totalCount++

		if len(response.Columns.ToWatch) >= limit &&
			len(response.Columns.InProgress) >= limit &&
			len(response.Columns.Watching) >= limit &&
			len(response.Columns.AwaitingRelease) >= limit &&
			len(response.Columns.Watched) >= limit {
			break
		}
	}

	response.Total = totalCount

	return response, nil
}

func (r *Repository) Update(id models.MaterialID, familyID models.FamilyID, profileID models.ProfileID, req *UpdateMaterialRequest) (*entities.Material, error) {
	sourceIDsJSON, err := json.Marshal(req.SourceIDs)
	if err != nil {
		return nil, fmt.Errorf("error marshaling source IDs: %v", err)
	}

	query := `
		UPDATE material
		SET name = $3, type = $4, runtime = $5, 
			poster_url = $6, source_ids = $7, classification_id = $8, watch_with = $9, status = $10, 
			priority = $11, notes = $12, updated_at = $13
		WHERE id = $1 AND family_id = $2 AND profile_id = $14 AND is_deleted = false`

	result, err := r.db.Exec(
		query, id, familyID, req.Name, req.Type,
		req.Runtime, req.PosterURL, sourceIDsJSON, req.ClassificationID, req.WatchWith, req.Status,
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

	return r.GetByID(id, profileID)
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

func (r *Repository) UpdateReview(id models.MaterialID, familyID models.FamilyID, profileID models.ProfileID, reviewScore float64) error {
	query := `
		UPDATE material
		SET review_score = $3, updated_at = $4
		WHERE id = $1 AND family_id = $2 AND profile_id = $5 AND is_deleted = false`

	result, err := r.db.Exec(query, id, familyID, reviewScore, time.Now().UTC(), profileID)
	if err != nil {
		return fmt.Errorf("error updating media review: %v", err)
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
