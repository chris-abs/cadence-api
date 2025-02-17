package user

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/storage/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	query := `
        INSERT INTO users (email, password, first_name, last_name, image_url, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`

	err = r.db.QueryRow(
		query,
		user.Email,
		string(hashedPassword),
		user.FirstName,
		user.LastName,
		user.ImageURL,
		time.Now().UTC(),
		time.Now().UTC(),
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
        WHERE email = $1`

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
        WHERE id = $1`

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

	containersQuery := `
        SELECT id, name, qr_code, qr_code_image, number, location, created_at, updated_at
        FROM container
        WHERE user_id = $1
        ORDER BY created_at DESC`

	rows, err := r.db.Query(containersQuery, id)
	if err != nil {
		return nil, fmt.Errorf("error getting containers: %v", err)
	}
	defer rows.Close()

	user.Containers = make([]models.Container, 0)
	for rows.Next() {
		var container models.Container
		err := rows.Scan(
			&container.ID,
			&container.Name,
			&container.QRCode,
			&container.QRCodeImage,
			&container.Number,
			&container.Location,
			&container.CreatedAt,
			&container.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning container: %v", err)
		}
		user.Containers = append(user.Containers, container)
	}

	return user, nil
}

func (r *Repository) GetAll() ([]*models.User, error) {
	query := `
        SELECT id, email, first_name, last_name, image_url, created_at, updated_at
        FROM users
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
        SET first_name = $2, last_name = $3, image_url = $4, updated_at = $5
        WHERE id = $1`

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

func (r *Repository) UpdateFamilyMembership(userID int, familyID *int, role models.UserRole) error {
    query := `
        UPDATE users
        SET family_id = $2,
            role = $3,
            updated_at = $4
        WHERE id = $1`

    result, err := r.db.Exec(
        query,
        userID,
        familyID,
        role,
        time.Now().UTC(),
    )

    if err != nil {
        return fmt.Errorf("error updating user family membership: %v", err)
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

func (r *Repository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
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
