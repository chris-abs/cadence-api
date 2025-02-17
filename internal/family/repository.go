package family

import (
	"database/sql"
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
    query := `
        INSERT INTO family (name, owner_id, module_permissions, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

    err := r.db.QueryRow(
        query,
        family.Name,
        family.OwnerID,
        family.ModulePermissions,
        time.Now().UTC(),
        time.Now().UTC(),
    ).Scan(&family.ID)

    if err != nil {
        return fmt.Errorf("error creating family: %v", err)
    }

    return nil
}

func (r *Repository) GetByID(id int) (*models.Family, error) {
    query := `
        SELECT id, name, owner_id, module_permissions, created_at, updated_at
        FROM family
        WHERE id = $1`

    family := new(models.Family)
    err := r.db.QueryRow(query, id).Scan(
        &family.ID,
        &family.Name,
        &family.OwnerID,
        &family.ModulePermissions,
        &family.CreatedAt,
        &family.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("family not found")
    }
    if err != nil {
        return nil, fmt.Errorf("error getting family: %v", err)
    }

    return family, nil
}

func (r *Repository) Update(family *models.Family) error {
    query := `
        UPDATE family
        SET name = $2, 
            module_permissions = $3,
            updated_at = $4
        WHERE id = $1`

    result, err := r.db.Exec(
        query,
        family.ID,
        family.Name,
        family.ModulePermissions,
        time.Now().UTC(),
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
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`

    err := r.db.QueryRow(
        query,
        invite.FamilyID,
        invite.Email,
        invite.Role,
        invite.Token,
        invite.ExpiresAt,
        time.Now().UTC(),
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
        WHERE token = $1`

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
        return nil, fmt.Errorf("invite not found")
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