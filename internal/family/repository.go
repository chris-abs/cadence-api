package family

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chrisabs/storage/internal/models"
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
	}

	family.Modules = defaultModules

	modulesJSON, err := json.Marshal(family.Modules)
	if err != nil {
		return fmt.Errorf("error marshaling modules: %v", err)
	}

	query := `
		INSERT INTO family (name, owner_id, modules, created_at, updated_at, status)
		VALUES ($1, $2, $3, $4, $4, $5)
		RETURNING id`

	err = r.db.QueryRow(
		query,
		family.Name,
		family.OwnerID,
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
		SELECT id, name, owner_id, modules, created_at, updated_at, status
		FROM family
		WHERE id = $1`

	var modulesJSON []byte
	family := new(models.Family)

	err := r.db.QueryRow(query, id).Scan(
		&family.ID,
		&family.Name,
		&family.OwnerID,
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
		WHERE id = $1`

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

func (r *Repository) CreateInvite(invite *models.FamilyInvite) error {
	query := `
		INSERT INTO family_invite (
			family_id, email, role, token, expires_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id`

	err := r.db.QueryRow(
		query,
		invite.FamilyID,
		invite.Email,
		invite.Role,
		invite.Token,
		invite.ExpiresAt,
		time.Now().UTC(),
	).Scan(&invite.ID)

	if err != nil {
		return fmt.Errorf("error creating family invite: %v", err)
	}

	return nil
}

func (r *Repository) GetInviteByToken(token string) (*models.FamilyInvite, error) {
	query := `
		SELECT id, family_id, email, role, token, expires_at, created_at, updated_at
		FROM family_invite
		WHERE token = $1 AND expires_at > NOW()`

	invite := new(models.FamilyInvite)
	err := r.db.QueryRow(query, token).Scan(
		&invite.ID,
		&invite.FamilyID,
		&invite.Email,
		&invite.Role,
		&invite.Token,
		&invite.ExpiresAt,
		&invite.CreatedAt,
		&invite.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invite not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting invite: %v", err)
	}

	return invite, nil
}

func (r *Repository) DeleteInvite(id int) error {
	query := `DELETE FROM family_invite WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting invite: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invite not found")
	}

	return nil
}
