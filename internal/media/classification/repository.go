package classification

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/media/classification/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(classification *entities.Classification) error {
	query := `
		INSERT INTO media_classification (
			id, name, description, color, image_url, family_id, profile_id, created_by, 
			created_at, updated_at, is_deleted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`
	
	now := time.Now().UTC()
	var createdAt, updatedAt time.Time
	
	err := r.db.QueryRow(
		query,
		classification.ID,
		classification.Name,
		classification.Description,
		classification.Color,
		classification.ImageURL,
		classification.FamilyID,
		classification.ProfileID,
		classification.CreatedBy,
		now,
		now,
		classification.IsDeleted,
	).Scan(&createdAt, &updatedAt)
	
	if err != nil {
		return fmt.Errorf("error creating classification: %v", err)
	}
	
	classification.CreatedAt = createdAt
	classification.UpdatedAt = updatedAt
	
	return nil
}

func (r *Repository) GetByID(id models.ClassificationID) (*entities.Classification, error) {
	query := `
		SELECT id, name, description, color, image_url, family_id, profile_id, created_by,
		       created_at, updated_at, is_deleted, deleted_at, deleted_by
		FROM media_classification 
		WHERE id = $1 AND is_deleted = false
	`
	
	var classification entities.Classification
	err := r.db.QueryRow(query, id).Scan(
		&classification.ID,
		&classification.Name,
		&classification.Description,
		&classification.Color,
		&classification.ImageURL,
		&classification.FamilyID,
		&classification.ProfileID,
		&classification.CreatedBy,
		&classification.CreatedAt,
		&classification.UpdatedAt,
		&classification.IsDeleted,
		&classification.DeletedAt,
		&classification.DeletedBy,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("classification not found")
		}
		return nil, fmt.Errorf("error getting classification: %v", err)
	}
	
	return &classification, nil
}

func (r *Repository) Update(classification *entities.Classification) error {
	query := `
		UPDATE media_classification 
		SET name = $2, description = $3, color = $4, image_url = $5, updated_at = $6
		WHERE id = $1 AND is_deleted = false
		RETURNING updated_at
	`
	
	var updatedAt time.Time
	err := r.db.QueryRow(
		query,
		classification.ID,
		classification.Name,
		classification.Description,
		classification.Color,
		classification.ImageURL,
		classification.UpdatedAt,
	).Scan(&updatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("classification not found or already deleted")
		}
		return fmt.Errorf("error updating classification: %v", err)
	}
	
	classification.UpdatedAt = updatedAt
	
	return nil
}

func (r *Repository) Delete(id models.ClassificationID, deletedBy models.ProfileID) error {
	query := `
		UPDATE media_classification 
		SET is_deleted = true, deleted_at = $2, deleted_by = $3
		WHERE id = $1 AND is_deleted = false
		RETURNING id
	`
	
	var deletedID models.ClassificationID
	err := r.db.QueryRow(query, id, time.Now().UTC(), deletedBy).Scan(&deletedID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("classification not found or already deleted")
		}
		return fmt.Errorf("error deleting classification: %v", err)
	}
	
	return nil
}

func (r *Repository) GetAllByProfile(profileID models.ProfileID, limit, offset int) ([]*entities.Classification, int, error) {
	countQuery := `
		SELECT COUNT(*) 
		FROM media_classification 
		WHERE profile_id = $1 AND is_deleted = false
	`
	
	var total int
	err := r.db.QueryRow(countQuery, profileID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting classifications: %v", err)
	}
	
	query := `
		SELECT id, name, description, color, image_url, family_id, profile_id, created_by,
		       created_at, updated_at, is_deleted, deleted_at, deleted_by
		FROM media_classification 
		WHERE profile_id = $1 AND is_deleted = false
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(query, profileID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting classifications: %v", err)
	}
	defer rows.Close()
	
	var classifications []*entities.Classification
	for rows.Next() {
		var classification entities.Classification
		err := rows.Scan(
			&classification.ID,
			&classification.Name,
			&classification.Description,
			&classification.Color,
			&classification.ImageURL,
			&classification.FamilyID,
			&classification.ProfileID,
			&classification.CreatedBy,
			&classification.CreatedAt,
			&classification.UpdatedAt,
			&classification.IsDeleted,
			&classification.DeletedAt,
			&classification.DeletedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning classification: %v", err)
		}
		
		classifications = append(classifications, &classification)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating classifications: %v", err)
	}
	
	return classifications, total, nil
}

func (r *Repository) GetMaterialCountByClassification(classificationID models.ClassificationID) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM material 
		WHERE classification_id = $1 AND is_deleted = false
	`
	
	var count int
	err := r.db.QueryRow(query, classificationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting material by classification: %v", err)
	}
	
	return count, nil
}