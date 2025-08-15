package classification

import (
	"database/sql"
	"fmt"

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
		INSERT INTO classifications (
			id, name, description, color, image_url, family_id, created_by, 
			created_at, updated_at, is_deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(query,
		classification.ID,
		classification.Name,
		classification.Description,
		classification.Color,
		classification.ImageURL,
		classification.FamilyID,
		classification.CreatedBy,
		classification.CreatedAt,
		classification.UpdatedAt,
		classification.IsDeleted,
	)
	
	if err != nil {
		return fmt.Errorf("error creating classification: %v", err)
	}
	
	return nil
}

func (r *Repository) GetByID(id models.ClassificationID) (*entities.Classification, error) {
	query := `
		SELECT id, name, description, color, image_url, family_id, created_by,
		       created_at, updated_at, is_deleted, deleted_at, deleted_by
		FROM classifications 
		WHERE id = ? AND is_deleted = false
	`
	
	var classification entities.Classification
	err := r.db.QueryRow(query, id).Scan(
		&classification.ID,
		&classification.Name,
		&classification.Description,
		&classification.Color,
		&classification.ImageURL,
		&classification.FamilyID,
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
		UPDATE classifications 
		SET name = ?, description = ?, color = ?, image_url = ?, updated_at = ?
		WHERE id = ? AND is_deleted = false
	`
	
	result, err := r.db.Exec(query,
		classification.Name,
		classification.Description,
		classification.Color,
		classification.ImageURL,
		classification.UpdatedAt,
		classification.ID,
	)
	
	if err != nil {
		return fmt.Errorf("error updating classification: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("classification not found or already deleted")
	}
	
	return nil
}

func (r *Repository) Delete(id models.ClassificationID, deletedBy models.ProfileID) error {
	query := `
		UPDATE classifications 
		SET is_deleted = true, deleted_at = NOW(), deleted_by = ?
		WHERE id = ? AND is_deleted = false
	`
	
	result, err := r.db.Exec(query, deletedBy, id)
	if err != nil {
		return fmt.Errorf("error deleting classification: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("classification not found or already deleted")
	}
	
	return nil
}

func (r *Repository) GetAllByFamily(familyID models.FamilyID, limit, offset int) ([]*entities.Classification, int, error) {
	countQuery := `
		SELECT COUNT(*) 
		FROM classifications 
		WHERE family_id = ? AND is_deleted = false
	`
	
	var total int
	err := r.db.QueryRow(countQuery, familyID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting classifications: %v", err)
	}
	
	query := `
		SELECT id, name, description, color, image_url, family_id, created_by,
		       created_at, updated_at, is_deleted, deleted_at, deleted_by
		FROM classifications 
		WHERE family_id = ? AND is_deleted = false
		ORDER BY name ASC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.db.Query(query, familyID, limit, offset)
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
		WHERE classification_id = ? AND is_deleted = false
	`
	
	var count int
	err := r.db.QueryRow(query, classificationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting material by classification: %v", err)
	}
	
	return count, nil
}
