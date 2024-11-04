package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type Handler struct {
	svc      *Service
	log      *zerolog.Logger
	validate *validator.Validate
}

func NewHandler(svc *Service, log *zerolog.Logger) *Handler {
	return &Handler{
		svc:      svc,
		log:      log,
		validate: validator.New(),
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.handleRegister)
	r.Post("/login", h.handleLogin)

	return r
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		switch err {
		case ErrUserExistsWithEmail:
			http.Error(w, err.Error(), http.StatusConflict)
		case ErrUserExistsWithUsername:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			h.log.Error().Err(err).Msg("registration failed")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			h.log.Error().Err(err).Msg("login failed")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
