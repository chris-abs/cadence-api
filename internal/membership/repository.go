package membership

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

func (r *Repository) Create(membership *models.FamilyMembership) error {
	query := `
		INSERT INTO family_membership (
			user_id, family_id, role, is_owner, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id`

	err := r.db.QueryRow(
		query,
		membership.UserID,
		membership.FamilyID,
		membership.Role,
		membership.IsOwner,
		time.Now().UTC(),
	).Scan(&membership.ID)

	if err != nil {
		return fmt.Errorf("error creating family membership: %v", err)
	}

	return nil
}

func (r *Repository) GetByID(id int) (*models.FamilyMembership, error) {
	query := `
		SELECT id, user_id, family_id, role, is_owner, created_at, updated_at
		FROM family_membership
		WHERE id = $1`

	membership := new(models.FamilyMembership)
	err := r.db.QueryRow(query, id).Scan(
		&membership.ID,
		&membership.UserID,
		&membership.FamilyID,
		&membership.Role,
		&membership.IsOwner,
		&membership.CreatedAt,
		&membership.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("membership not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting membership: %v", err)
	}

	return membership, nil
}

func (r *Repository) GetByUserID(userID int) ([]*models.FamilyMembership, error) {
	query := `
		SELECT id, user_id, family_id, role, is_owner, created_at, updated_at
		FROM family_membership
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting memberships: %v", err)
	}
	defer rows.Close()

	var memberships []*models.FamilyMembership
	for rows.Next() {
		membership := new(models.FamilyMembership)
		err := rows.Scan(
			&membership.ID,
			&membership.UserID,
			&membership.FamilyID,
			&membership.Role,
			&membership.IsOwner,
			&membership.CreatedAt,
			&membership.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning membership: %v", err)
		}
		memberships = append(memberships, membership)
	}

	return memberships, nil
}

func (r *Repository) GetActiveMembershipForUser(userID int) (*models.FamilyMembership, error) {
	query := `
		SELECT id, user_id, family_id, role, is_owner, created_at, updated_at
		FROM family_membership
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	membership := new(models.FamilyMembership)
	err := r.db.QueryRow(query, userID).Scan(
		&membership.ID,
		&membership.UserID,
		&membership.FamilyID,
		&membership.Role,
		&membership.IsOwner,
		&membership.CreatedAt,
		&membership.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no membership found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting membership: %v", err)
	}

	return membership, nil
}

func (r *Repository) GetByFamilyID(familyID int) ([]*models.FamilyMembership, error) {
	query := `
		SELECT id, user_id, family_id, role, is_owner, created_at, updated_at
		FROM family_membership
		WHERE family_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, familyID)
	if err != nil {
		return nil, fmt.Errorf("error getting memberships: %v", err)
	}
	defer rows.Close()

	var memberships []*models.FamilyMembership
	for rows.Next() {
		membership := new(models.FamilyMembership)
		err := rows.Scan(
			&membership.ID,
			&membership.UserID,
			&membership.FamilyID,
			&membership.Role,
			&membership.IsOwner,
			&membership.CreatedAt,
			&membership.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning membership: %v", err)
		}
		memberships = append(memberships, membership)
	}

	return memberships, nil
}

func (r *Repository) GetFamilyOwner(familyID int) (*models.FamilyMembership, error) {
	query := `
		SELECT id, user_id, family_id, role, is_owner, created_at, updated_at
		FROM family_membership
		WHERE family_id = $1 AND is_owner = true
		LIMIT 1`

	membership := new(models.FamilyMembership)
	err := r.db.QueryRow(query, familyID).Scan(
		&membership.ID,
		&membership.UserID,
		&membership.FamilyID,
		&membership.Role,
		&membership.IsOwner,
		&membership.CreatedAt,
		&membership.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no owner found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting owner: %v", err)
	}

	return membership, nil
}

func (r *Repository) Update(membership *models.FamilyMembership) error {
	query := `
		UPDATE family_membership
		SET role = $2, 
			is_owner = $3,
			updated_at = $4
		WHERE id = $1`

	result, err := r.db.Exec(
		query,
		membership.ID,
		membership.Role,
		membership.IsOwner,
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("error updating membership: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("membership not found")
	}

	return nil
}

func (r *Repository) Delete(id int) error {
	query := `DELETE FROM family_membership WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting membership: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("membership not found")
	}

	return nil
}