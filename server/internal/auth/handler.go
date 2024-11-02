package auth

import (
	"discord/internal/http/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type Handler struct {
	auth   AuthService
	tokens TokenService
}

func NewHandler(auth AuthService, tokens TokenService) *Handler {
	return &Handler{
		auth:   auth,
		tokens: tokens,
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Render(w, r, response.ErrInvalidRequest(err))
		return
	}

	user, err := h.auth.Login(r.Context(), req)
	if err != nil {
		render.Render(w, r, response.ErrUnauthorized())
		return
	}

	token, err := h.tokens.Generate(user)
	if err != nil {
		render.Render(w, r, response.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Render(w, r, response.ErrInvalidRequest(err))
		return
	}

	user, err := h.auth.Register(r.Context(), req)
	if err != nil {
		render.Render(w, r, response.ErrConflict("User already exists"))
		return
	}

	token, err := h.tokens.Generate(user)
	if err != nil {
		render.Render(w, r, response.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	return r
}
