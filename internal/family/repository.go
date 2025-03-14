package family

import (
	"database/sql"
	"encoding/json"
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

func (r *Repository) Create(family *models.Family) error {
	defaultModules := []models.Module{
		{
			ID: "storage",
			IsEnabled: true,
		},
		{
			ID: "meals",
			IsEnabled: false,
		},
		{
			ID: "chores",
			IsEnabled: false,
		},
		{
			ID: "services",
			IsEnabled: false,
		},
	}

	family.Modules = defaultModules

	modulesJSON, err := json.Marshal(family.Modules)
	if err != nil {
		return fmt.Errorf("error marshaling modules: %v", err)
	}

	query := `
		INSERT INTO family (name, modules, created_at, updated_at, status)
		VALUES ($1, $2, $3, $3, $4)
		RETURNING id`

	err = r.db.QueryRow(
		query,
		family.Name,
		modulesJSON,
		time.Now().UTC(),
		models.FamilyStatusActive,
	).Scan(&family.ID)

	if err != nil {
		return fmt.Errorf("error creating family: %v", err)
	}

	return nil
}

func (r *Repository) GetByID(id int) (*models.Family, error) {
    query := `
        SELECT id, name, modules, created_at, updated_at, status
        FROM family
        WHERE id = $1 AND is_deleted = false`

	var modulesJSON []byte
	family := new(models.Family)

	err := r.db.QueryRow(query, id).Scan(
		&family.ID,
		&family.Name,
		&modulesJSON,
		&family.CreatedAt,
		&family.UpdatedAt,
		&family.Status,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("family not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting family: %v", err)
	}

	if err := json.Unmarshal(modulesJSON, &family.Modules); err != nil {
		return nil, fmt.Errorf("error unmarshaling modules: %v", err)
	}

	return family, nil
}

func (r *Repository) Update(family *models.Family) error {
    modulesJSON, err := json.Marshal(family.Modules)
    if err != nil {
        return fmt.Errorf("error marshaling modules: %v", err)
    }

    query := `
        UPDATE family
        SET name = $2,
            modules = $3,
            updated_at = $4,
            status = $5
        WHERE id = $1 AND is_deleted = false`

	result, err := r.db.Exec(
		query,
		family.ID,
		family.Name,
		modulesJSON,
		time.Now().UTC(),
		family.Status,
	)

	if err != nil {
		return fmt.Errorf("error updating family: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family not found")
	}

	return nil
}

func (r *Repository) Delete(id int, deletedBy int) error {
    query := `
        UPDATE family
        SET is_deleted = true, deleted_at = $2, deleted_by = $3, updated_at = $2
        WHERE id = $1 AND is_deleted = false`
    
    result, err := r.db.Exec(query, id, time.Now().UTC(), deletedBy)
    if err != nil {
        return fmt.Errorf("error soft deleting family: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking delete result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("family not found")
    }

    return nil
}

func (r *Repository) RestoreFamily(id int) error {
    query := `
        UPDATE family
        SET is_deleted = false, deleted_at = NULL, deleted_by = NULL, updated_at = $2
        WHERE id = $1 AND is_deleted = true`
    
    result, err := r.db.Exec(query, id, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error restoring family: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking restore result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("family not found or not deleted")
    }

    return nil
}