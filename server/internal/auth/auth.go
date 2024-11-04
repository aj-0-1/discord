package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username" validate:"required,min=3,max=30"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type Service struct {
	db  *sql.DB
	key []byte
	log *zerolog.Logger
}

var (
	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrUserExistsWithEmail    = errors.New("user with this email already exists")
	ErrUserExistsWithUsername = errors.New("user with this username already exists")
	ErrInvalidToken           = errors.New("invalid or expired token")
)

func isPgUniqueViolation(err error) bool {
	pgErr, ok := err.(*pq.Error)
	return ok && pgErr.Code == "23505" // unique_violation
}

func isEmailConstraint(err error) bool {
	pgErr, ok := err.(*pq.Error)
	return ok && pgErr.Constraint == "users_email_key"
}

func isUsernameConstraint(err error) bool {
	pgErr, ok := err.(*pq.Error)
	return ok && pgErr.Constraint == "users_username_key"
}

func NewService(db *sql.DB, jwtKey []byte, log *zerolog.Logger) *Service {
	return &Service{
		db:  db,
		key: jwtKey,
		log: log,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	user := &User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	const q = `
        INSERT INTO users (id, email, username, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, email, username, created_at, updated_at`

	err = tx.QueryRowContext(ctx, q,
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
	if err != nil {
		if isPgUniqueViolation(err) {
			if isEmailConstraint(err) {
				return nil, ErrUserExistsWithEmail
			}
			if isUsernameConstraint(err) {
				return nil, ErrUserExistsWithUsername
			}
			return nil, fmt.Errorf("insert user: %w", err)
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	token, err := s.createToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	var user User
	const q = `
        SELECT id, email, username, password_hash, created_at, updated_at
        FROM users
        WHERE email = $1`

	err := s.db.QueryRowContext(ctx, q, req.Email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.createToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User:  &user,
	}, nil
}

func (s *Service) VerifyToken(token string) (string, error) {
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.key, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return "", ErrInvalidToken
		}
		return "", fmt.Errorf("parse token: %w", err)
	}

	if !tkn.Valid {
		return "", ErrInvalidToken
	}

	return claims.UserID, nil
}

func (s *Service) createToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.key)
}
