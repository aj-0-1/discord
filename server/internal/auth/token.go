package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type tokenService struct {
    secretKey []byte
    duration  time.Duration
}

func NewTokenService(secretKey string, duration time.Duration) TokenService {
    return &tokenService{
        secretKey: []byte(secretKey),
        duration:  duration,
    }
}

type Claims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}

func (s *tokenService) Generate(user *User) (string, error) {
    now := time.Now()
    claims := Claims{
        UserID: user.ID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(s.duration)),
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secretKey)
}

func (s *tokenService) Validate(tokenString string) (string, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &Claims{},
        func(token *jwt.Token) (interface{}, error) {
            return s.secretKey, nil
        },
    )
    if err != nil {
        return "", err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims.UserID, nil
    }

    return "", jwt.ErrSignatureInvalid
}
