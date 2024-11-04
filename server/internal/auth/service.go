package auth

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (*User, error)
	Register(ctx context.Context, req RegisterRequest) (*User, error)
}

type TokenService interface {
	Generate(user *User) (string, error)
	Validate(token string) (string, error) // returns userID
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type authService struct {
	users  UserRepository
	tokens TokenService
}

func NewAuthService(users UserRepository, tokens TokenService) AuthService {
	return &authService{
		users:  users,
		tokens: tokens,
	}
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*User, error) {
	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	exists, err := s.users.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExistsWithEmail
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
