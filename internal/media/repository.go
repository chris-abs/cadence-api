package media

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chrisabs/cadence/internal/media/entities"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(profileID int, familyID int, req *CreateMediaRequest) (*entities.Media, error) {
	query := `
		INSERT INTO media (
			name, type, genre, release_year, runtime, poster_url, source_ids,
			watch_with, status, priority, notes, profile_id, family_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`

	sourceIDsJSON, err := json.Marshal(req.SourceIDs)
	if err != nil {
		return nil, fmt.Errorf("error marshaling source IDs: %v", err)
	}

	now := time.Now().UTC()
	var mediaID int

	err = r.db.QueryRow(
		query,
		req.Name, req.Type, req.Genre, req.ReleaseYear, req.Runtime, req.PosterURL,
		sourceIDsJSON, req.WatchWith, req.Status, req.Priority, req.Notes,
		profileID, familyID, now, now,
	).Scan(&mediaID)

	if err != nil {
		return nil, fmt.Errorf("error creating media: %v", err)
	}

	return r.GetByID(mediaID, familyID)
}

func (r *Repository) GetByID(id int, familyID int) (*entities.Media, error) {
	query := `
		SELECT m.id, m.name, m.type, m.genre, m.release_year, m.runtime, m.poster_url,
			   m.source_ids, m.watch_with, m.status, m.priority, m.notes,
			   m.profile_id, m.family_id, m.created_at, m.updated_at
		FROM media m
		WHERE m.id = $1 AND m.family_id = $2 AND m.is_deleted = false`

	media := new(entities.Media)
	var sourceIDsJSON []byte

	err := r.db.QueryRow(query, id, familyID).Scan(
		&media.ID, &media.Name, &media.Type, &media.Genre, &media.ReleaseYear,
		&media.Runtime, &media.PosterURL, &sourceIDsJSON, &media.WatchWith,
		&media.Status, &media.Priority, &media.Notes, &media.ProfileID,
		&media.FamilyID, &media.CreatedAt, &media.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("media not found")
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(sourceIDsJSON, &media.SourceIDs); err != nil {
		return nil, fmt.Errorf("error parsing source IDs: %v", err)
	}

	if err := r.loadMediaSources(media); err != nil {
		return nil, err
	}

	return media, nil
}

func (r *Repository) Search(familyID int, req *MediaSearchRequest) ([]entities.Media, int, error) {
	profileID := req.ProfileID
	if profileID == nil {
		return nil, 0, fmt.Errorf("profile ID is required")
	}

	quickCheckQuery := `
		SELECT EXISTS (
			SELECT 1 FROM media 
			WHERE family_id = $1 AND profile_id = $2 AND is_deleted = false
		)`
	
	var hasResults bool
	err := r.db.QueryRow(quickCheckQuery, familyID, *profileID).Scan(&hasResults)
	if err != nil {
		return nil, 0, err
	}

	if !hasResults {
		return []entities.Media{}, 0, nil
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

	if req.SourceID > 0 {
		conditions = append(conditions, fmt.Sprintf("m.source_ids ? $%d", argIndex))
		args = append(args, fmt.Sprintf("%d", req.SourceID))
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

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM media m WHERE %s", whereClause)
	var total int
	err = r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
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
			FROM media m
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

	var mediaList []entities.Media
	for rows.Next() {
		var media entities.Media
		var sourceIDsJSON []byte
		var rank int

		err := rows.Scan(
			&media.ID, &media.Name, &media.Type, &media.Genre, &media.ReleaseYear,
			&media.Runtime, &media.PosterURL, &sourceIDsJSON, &media.WatchWith,
			&media.Status, &media.Priority, &media.Notes, &media.ProfileID,
			&media.FamilyID, &media.CreatedAt, &media.UpdatedAt, &rank,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning media result: %v", err)
		}

		if err := json.Unmarshal(sourceIDsJSON, &media.SourceIDs); err != nil {
			return nil, 0, fmt.Errorf("error parsing source IDs: %v", err)
		}

		if err := r.loadMediaSources(&media); err != nil {
			return nil, 0, fmt.Errorf("error loading sources: %v", err)
		}

		mediaList = append(mediaList, media)
	}

	return mediaList, total, nil
}

func (r *Repository) Update(id int, familyID int, profileID int, req *UpdateMediaRequest) (*entities.Media, error) {
	sourceIDsJSON, err := json.Marshal(req.SourceIDs)
	if err != nil {
		return nil, fmt.Errorf("error marshaling source IDs: %v", err)
	}

	query := `
		UPDATE media
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

func (r *Repository) UpdateStatus(id int, familyID int, profileID int, status entities.Status) error {
	query := `
		UPDATE media
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

func (r *Repository) Delete(id int, familyID int, profileID int, deletedBy int) error {
	query := `
		UPDATE media
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

func (r *Repository) GetStatusSummary(familyID int, profileID int) (*MediaStatusSummaryResponse, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN status = 'to_watch' THEN 1 END) as to_watch,
			COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress,
			COUNT(CASE WHEN status = 'watching' THEN 1 END) as watching,
			COUNT(CASE WHEN status = 'awaiting_release' THEN 1 END) as awaiting_release,
			COUNT(CASE WHEN status = 'watched' THEN 1 END) as watched,
			COUNT(*) as total
		FROM media
		WHERE family_id = $1 AND profile_id = $2 AND is_deleted = false`

	summary := new(MediaStatusSummaryResponse)
	err := r.db.QueryRow(query, familyID, profileID).Scan(
		&summary.ToWatch, &summary.InProgress, &summary.Watching,
		&summary.AwaitingRelease, &summary.Watched, &summary.Total,
	)

	if err != nil {
		return nil, fmt.Errorf("error getting status summary: %v", err)
	}

	return summary, nil
}

func (r *Repository) loadMediaSources(media *entities.Media) error {
	if len(media.SourceIDs) == 0 {
		media.Sources = make([]entities.Source, 0)
		return nil
	}

	placeholders := make([]string, len(media.SourceIDs))
	args := make([]interface{}, len(media.SourceIDs))
	for i, sourceID := range media.SourceIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = sourceID
	}

	query := fmt.Sprintf(`
		SELECT id, name, color, category
		FROM media_source
		WHERE id IN (%s) AND is_deleted = false
		ORDER BY name`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	media.Sources = make([]entities.Source, 0)
	for rows.Next() {
		var source entities.Source
		err := rows.Scan(&source.ID, &source.Name, &source.Color, &source.Category)
		if err != nil {
			return err
		}
		media.Sources = append(media.Sources, source)
	}

	return nil
}

func (r *Repository) GetAllSources() ([]entities.Source, error) {
	query := `
		SELECT id, name, color, category
		FROM media_source
		WHERE is_deleted = false
		ORDER BY category, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []entities.Source
	for rows.Next() {
		var source entities.Source
		err := rows.Scan(&source.ID, &source.Name, &source.Color, &source.Category)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}

	return sources, nil
}

func (r *Repository) SeedSources() error {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM media_source WHERE is_deleted = false").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	query := `
		INSERT INTO media_source (name, color, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	now := time.Now().UTC()
	for _, source := range entities.DefaultSources {
		_, err := r.db.Exec(query, source.Name, source.Color, source.Category, now, now)
		if err != nil {
			return fmt.Errorf("error seeding source %s: %v", source.Name, err)
		}
	}

	return nil
}