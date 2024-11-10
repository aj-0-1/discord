package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/rs/zerolog"
)

type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

type Service struct {
	db  *sql.DB
	log *zerolog.Logger
}

func NewService(db *sql.DB, log *zerolog.Logger) *Service {
	return &Service{
		db:  db,
		log: log,
	}
}

// GetByEmail used by auth service
func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	const q = `
        SELECT id, email, username, password_hash, created_at, updated_at
        FROM users
        WHERE email = $1`

	err := s.db.QueryRowContext(ctx, q, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create used by auth service during registration
func (s *Service) Create(ctx context.Context, user *User) error {
	const q = `
        INSERT INTO users (id, email, username, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, email, username, created_at, updated_at`

	return s.db.QueryRowContext(ctx, q,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}

// SearchUsers for the chat feature
func (s *Service) SearchUsers(ctx context.Context, query string, excludeUserID string) ([]User, error) {
	const q = `
        SELECT id, email, username, created_at, updated_at
        FROM users 
        WHERE 
            id != $1 AND
            (
                username ILIKE $2 OR 
                email ILIKE $2
            )
        LIMIT 10`

	rows, err := s.db.QueryContext(ctx, q, excludeUserID, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
