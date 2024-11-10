package user

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Handler struct {
	svc *Service
	log *zerolog.Logger
}

func NewHandler(svc *Service, log *zerolog.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

// @Summary Search users
// @Description Search users by username or email
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param q query string true "Search query"
// @Success 200 {array} User
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Router /users/search [get]
func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "search query is required", http.StatusBadRequest)
		return
	}

	// Get current user ID from context (set by auth middleware)
	userID := r.Context().Value("userID")
	if userID == nil {
		h.log.Warn().Msg("userID is missing or invalid in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	parsedUserID := userID.(uuid.UUID)

	users, err := h.svc.SearchUsers(r.Context(), query, parsedUserID.String())
	if err != nil {
		h.log.Error().Err(err).
			Str("query", query).
			Msg("failed to search users")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		h.log.Error().Err(err).Msg("failed to encode response")
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/search", h.handleSearch)

	return r
}
