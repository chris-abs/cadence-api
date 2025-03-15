package profile

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(profile *models.Profile) error {
	query := `
		INSERT INTO profile (
			family_id, name, role, pin, image_url, is_owner, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		profile.FamilyID,
		profile.Name,
		profile.Role,
		profile.Pin,
		profile.ImageURL,
		profile.IsOwner,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&profile.ID, &profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating profile: %v", err)
	}

	return nil
}

func (r *Repository) GetByID(id int) (*models.Profile, error) {
	query := `
		SELECT id, family_id, name, role, pin, image_url, is_owner, created_at, updated_at
		FROM profile
		WHERE id = $1 AND is_deleted = false`

	profile := new(models.Profile)
	err := r.db.QueryRow(query, id).Scan(
		&profile.ID,
		&profile.FamilyID,
		&profile.Name,
		&profile.Role,
		&profile.Pin,
		&profile.ImageURL,
		&profile.IsOwner,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("profile not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting profile: %v", err)
	}

	return profile, nil
}

func (r *Repository) GetByFamilyID(familyID int) ([]*models.Profile, error) {
	query := `
		SELECT id, family_id, name, role, pin, image_url, is_owner, created_at, updated_at
		FROM profile
		WHERE family_id = $1 AND is_deleted = false
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, familyID)
	if err != nil {
		return nil, fmt.Errorf("error querying profiles: %v", err)
	}
	defer rows.Close()

	var profiles []*models.Profile
	for rows.Next() {
		profile := new(models.Profile)
		err := rows.Scan(
			&profile.ID,
			&profile.FamilyID,
			&profile.Name,
			&profile.Role,
			&profile.Pin,
			&profile.ImageURL,
			&profile.IsOwner,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning profile: %v", err)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (r *Repository) Update(profile *models.Profile) error {
	query := `
		UPDATE profile
		SET name = $2, 
			role = $3, 
			pin = $4,
			image_url = $5,
			is_owner = $6,
			updated_at = $7
		WHERE id = $1 AND family_id = $8 AND is_deleted = false`

	result, err := r.db.Exec(
		query,
		profile.ID,
		profile.Name,
		profile.Role,
		profile.Pin,
		profile.ImageURL,
		profile.IsOwner,
		time.Now().UTC(),
		profile.FamilyID,
	)

	if err != nil {
		return fmt.Errorf("error updating profile: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("profile not found or access denied")
	}

	return nil
}

func (r *Repository) Delete(id int, familyID int, deletedBy int) error {
	query := `
		UPDATE profile 
		SET is_deleted = true, deleted_at = $3, deleted_by = $4, updated_at = $3
		WHERE id = $1 AND family_id = $2 AND is_deleted = false`
	
	result, err := r.db.Exec(query, id, familyID, time.Now().UTC(), deletedBy)
	if err != nil {
		return fmt.Errorf("error soft deleting profile: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("profile not found or access denied")
	}

	return nil
}

func (r *Repository) Restore(id int, familyID int) error {
	query := `
		UPDATE profile
		SET is_deleted = false, deleted_at = NULL, deleted_by = NULL, updated_at = $3
		WHERE id = $1 AND family_id = $2 AND is_deleted = true`
	
	result, err := r.db.Exec(query, id, familyID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("error restoring profile: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking restore result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("profile not found or not deleted")
	}

	return nil
}

func (r *Repository) GetOwnerProfile(familyID int) (*models.Profile, error) {
	query := `
		SELECT id, family_id, name, role, pin, image_url, is_owner, created_at, updated_at
		FROM profile
		WHERE family_id = $1 AND is_owner = true AND is_deleted = false
		LIMIT 1`

	profile := new(models.Profile)
	err := r.db.QueryRow(query, familyID).Scan(
		&profile.ID,
		&profile.FamilyID,
		&profile.Name,
		&profile.Role,
		&profile.Pin,
		&profile.ImageURL,
		&profile.IsOwner,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("owner profile not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting owner profile: %v", err)
	}

	return profile, nil
}