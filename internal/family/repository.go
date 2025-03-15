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

func (r *Repository) Create(family *FamilyAccount) error {
	query := `
		INSERT INTO family_account (email, password, family_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		family.Email,
		family.Password,
		family.FamilyName,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&family.ID, &family.CreatedAt, &family.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating family account: %v", err)
	}

	return nil
}

func (r *Repository) CreateSettings(settings *FamilySettings) error {
	modulesJSON, err := json.Marshal(settings.Modules)
	if err != nil {
		return fmt.Errorf("error marshaling modules: %v", err)
	}

	query := `
		INSERT INTO family_settings (family_id, modules, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`

	err = r.db.QueryRow(
		query,
		settings.FamilyID,
		modulesJSON,
		settings.Status,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating family settings: %v", err)
	}

	return nil
}

func (r *Repository) GetByID(id int) (*FamilyAccount, error) {
	query := `
		SELECT id, email, password, family_name, created_at, updated_at
		FROM family_account
		WHERE id = $1 AND is_deleted = false`

	family := new(FamilyAccount)
	err := r.db.QueryRow(query, id).Scan(
		&family.ID,
		&family.Email,
		&family.Password,
		&family.FamilyName,
		&family.CreatedAt,
		&family.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("family account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting family account: %v", err)
	}

	return family, nil
}

func (r *Repository) GetByEmail(email string) (*FamilyAccount, error) {
	query := `
		SELECT id, email, password, family_name, created_at, updated_at
		FROM family_account
		WHERE email = $1 AND is_deleted = false`

	family := new(FamilyAccount)
	err := r.db.QueryRow(query, email).Scan(
		&family.ID,
		&family.Email,
		&family.Password,
		&family.FamilyName,
		&family.CreatedAt,
		&family.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("family account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting family account: %v", err)
	}

	return family, nil
}

func (r *Repository) GetSettings(familyID int) (*FamilySettings, error) {
	query := `
		SELECT family_id, modules, status, created_at, updated_at
		FROM family_settings
		WHERE family_id = $1 AND is_deleted = false`

	settings := new(FamilySettings)
	var modulesJSON []byte

	err := r.db.QueryRow(query, familyID).Scan(
		&settings.FamilyID,
		&modulesJSON,
		&settings.Status,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("family settings not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting family settings: %v", err)
	}

	if err := json.Unmarshal(modulesJSON, &settings.Modules); err != nil {
		return nil, fmt.Errorf("error unmarshaling modules: %v", err)
	}

	return settings, nil
}

func (r *Repository) Update(family *FamilyAccount) error {
	query := `
		UPDATE family_account
		SET family_name = $2, updated_at = $3
		WHERE id = $1 AND is_deleted = false`

	result, err := r.db.Exec(
		query,
		family.ID,
		family.FamilyName,
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("error updating family account: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family account not found")
	}

	return nil
}

func (r *Repository) UpdateSettings(settings *FamilySettings) error {
	modulesJSON, err := json.Marshal(settings.Modules)
	if err != nil {
		return fmt.Errorf("error marshaling modules: %v", err)
	}

	query := `
		UPDATE family_settings
		SET modules = $2, status = $3, updated_at = $4
		WHERE family_id = $1 AND is_deleted = false`

	result, err := r.db.Exec(
		query,
		settings.FamilyID,
		modulesJSON,
		settings.Status,
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("error updating family settings: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family settings not found")
	}

	return nil
}

func (r *Repository) UpdateModule(familyID int, moduleID models.ModuleID, isEnabled bool) error {
	settings, err := r.GetSettings(familyID)
	if err != nil {
		return err
	}

	moduleFound := false
	for i, module := range settings.Modules {
		if module.ID == moduleID {
			settings.Modules[i].IsEnabled = isEnabled
			moduleFound = true
			break
		}
	}

	if !moduleFound {
		settings.Modules = append(settings.Modules, models.Module{
			ID:        moduleID,
			IsEnabled: isEnabled,
		})
	}

	return r.UpdateSettings(settings)
}

func (r *Repository) Delete(id int, deletedBy int) error {
	query := `
		UPDATE family_account
		SET is_deleted = true, deleted_at = $2, deleted_by = $3, updated_at = $2
		WHERE id = $1 AND is_deleted = false`

	result, err := r.db.Exec(query, id, time.Now().UTC(), deletedBy)
	if err != nil {
		return fmt.Errorf("error deleting family account: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family account not found")
	}

	return nil
}

func (r *Repository) Restore(id int) error {
	query := `
		UPDATE family_account
		SET is_deleted = false, deleted_at = NULL, deleted_by = NULL, updated_at = $2
		WHERE id = $1 AND is_deleted = true`

	result, err := r.db.Exec(query, id, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("error restoring family account: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking restore result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family account not found or not deleted")
	}

	return nil
}

func (r *Repository) IsModuleEnabled(familyID int, moduleID models.ModuleID) (bool, error) {
	settings, err := r.GetSettings(familyID)
	if err != nil {
		return false, err
	}

	for _, module := range settings.Modules {
		if module.ID == moduleID {
			return module.IsEnabled, nil
		}
	}

	return false, nil
}