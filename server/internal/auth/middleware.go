package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		userID, err := s.VerifyToken(parts[1])
		if err != nil {
			s.log.Print("Failed to verify token")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parsedID, err := uuid.Parse(userID)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", parsedID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestLogger(log *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				log.Info().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("requestId", middleware.GetReqID(r.Context())).
					Int("status", ww.Status()).
					Int("bytes", ww.BytesWritten()).
					Dur("duration", time.Since(start)).
					Str("ip", r.RemoteAddr).
					Str("user-agent", r.UserAgent()).
					Msg("request completed")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
