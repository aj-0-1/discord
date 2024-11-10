package auth

import (
	"context"
	"database/sql"
	"discord/internal/user"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// @title Discord API
// @version 1.0
// @description API for discord clone
// @host localhost:8080
// @BasePath /api

// LoginRequest represents login credentials
// @Description Login request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Invalid credentials"
// @Router /auth/login [post]
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username" validate:"required,min=3,max=30"`
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  *user.User `json:"user"`
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type Service struct {
	userService *user.Service
	key         []byte
	log         *zerolog.Logger
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

func NewService(userService *user.Service, jwtKey []byte, log *zerolog.Logger) *Service {
	return &Service{
		userService: userService,
		key:         jwtKey,
		log:         log,
	}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userService.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
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
		User:  user,
	}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &user.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userService.Create(ctx, user); err != nil {
		if isPgUniqueViolation(err) {
			if isEmailConstraint(err) {
				return nil, ErrUserExistsWithEmail
			}
			if isUsernameConstraint(err) {
				return nil, ErrUserExistsWithUsername
			}
		}
		return nil, fmt.Errorf("create user: %w", err)
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
