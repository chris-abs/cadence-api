package user

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) BeginTx() (*sql.Tx, error) {
    return r.db.Begin()
}

func (r *Repository) CreateTx(tx *sql.Tx, user *models.User) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("error hashing password: %v", err)
    }

    query := `
        INSERT INTO users (email, password, first_name, last_name, image_url, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`

    err = tx.QueryRow(
        query,
        user.Email,
        string(hashedPassword),
        user.FirstName,
        user.LastName,
        user.ImageURL,
        user.CreatedAt,
        user.UpdatedAt,
    ).Scan(&user.ID)

    if err != nil {
        return fmt.Errorf("error creating user: %v", err)
    }

    return nil
}

func (r *Repository) GetByEmail(email string) (*models.User, error) {
    query := `
        SELECT id, email, password, first_name, last_name, image_url, created_at, updated_at
        FROM users
        WHERE email = $1 AND is_deleted = false`

    user := new(models.User)
    err := r.db.QueryRow(query, email).Scan(
        &user.ID,
        &user.Email,
        &user.Password,
        &user.FirstName,
        &user.LastName,
        &user.ImageURL,
        &user.CreatedAt,
        &user.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    if err != nil {
        return nil, fmt.Errorf("error getting user: %v", err)
    }

    return user, nil
}

func (r *Repository) GetByID(id int) (*models.User, error) {
    query := `
        SELECT id, email, first_name, last_name, image_url, created_at, updated_at
        FROM users
        WHERE id = $1 AND is_deleted = false`

    user := new(models.User)
    err := r.db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Email,
        &user.FirstName,
        &user.LastName,
        &user.ImageURL,
        &user.CreatedAt,
        &user.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    if err != nil {
        return nil, fmt.Errorf("error getting user: %v", err)
    }

    return user, nil
}

func (r *Repository) GetAll() ([]*models.User, error) {
    query := `
        SELECT id, email, first_name, last_name, image_url, created_at, updated_at
        FROM users
        WHERE is_deleted = false
        ORDER BY created_at DESC`

    rows, err := r.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("error getting users: %v", err)
    }
    defer rows.Close()

    var users []*models.User
    for rows.Next() {
        user := new(models.User)
        err := rows.Scan(
            &user.ID,
            &user.Email,
            &user.FirstName,
            &user.LastName,
            &user.ImageURL,
            &user.CreatedAt,
            &user.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning user: %v", err)
        }
        users = append(users, user)
    }

    return users, nil
}

func (r *Repository) Update(user *models.User) error {
    query := `
        UPDATE users
        SET first_name = $2, 
            last_name = $3, 
            image_url = $4,
            updated_at = $5
        WHERE id = $1 AND is_deleted = false`

    result, err := r.db.Exec(
        query,
        user.ID,
        user.FirstName,
        user.LastName,
        user.ImageURL,
        time.Now().UTC(),
    )

    if err != nil {
        return fmt.Errorf("error updating user: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking update result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("user not found")
    }

    return nil
}

func (r *Repository) Delete(id int, deletedBy int) error {
    query := `UPDATE users 
              SET is_deleted = true, deleted_at = $2, deleted_by = $3, updated_at = $2
              WHERE id = $1 AND is_deleted = false`
    
    result, err := r.db.Exec(query, id, time.Now().UTC(), deletedBy)
    if err != nil {
        return fmt.Errorf("error soft deleting user: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking delete result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("user not found")
    }

    return nil
}

func (r *Repository) RestoreUser(id int) error {
    query := `
        UPDATE users
        SET is_deleted = false, deleted_at = NULL, deleted_by = NULL, updated_at = $2
        WHERE id = $1 AND is_deleted = true`
    
    result, err := r.db.Exec(query, id, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error restoring user: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking restore result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("user not found or not deleted")
    }

    return nil
}