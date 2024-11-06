package chat

import (
	"discord/internal/auth"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type Handler struct {
	svc      *Service
	log      *zerolog.Logger
	upgrader websocket.Upgrader
}

func NewHandler(svc *Service, log *zerolog.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// TODO: In production, implement proper origin checking
				return true
			},
		},
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/messages", h.handleSendMessage)
	r.Get("/messages/{userID}", h.handleGetMessages) // This line needed fixing
	r.Get("/ws", h.handleWebSocket)

	return r
}

func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// The auth middleware should have already verified the token and added userID to context
	userID, ok := r.Context().Value(auth.KeyUserID).(uuid.UUID)
	if !ok {
		h.log.Error().Msg("user ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.log.Debug().
		Str("userId", userID.String()).
		Str("remoteAddr", r.RemoteAddr).
		Msg("websocket connection attempt")

	// Upgrade connection
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to upgrade connection")
		return
	}

	client := &Client{
		hub:    h.svc.hub,
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
	}

	h.log.Info().
		Str("userId", userID.String()).
		Str("remoteAddr", conn.RemoteAddr().String()).
		Msg("new websocket connection established")

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (h *Handler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var msg struct {
		ToID    string `json:"toId"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Get sender ID from context as UUID
	fromID, ok := r.Context().Value(auth.KeyUserID).(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse toID into UUID
	toID, err := uuid.Parse(msg.ToID)
	if err != nil {
		http.Error(w, "invalid recipient id", http.StatusBadRequest)
		return
	}

	message := &Message{
		ID:      uuid.New(),
		FromID:  fromID,
		ToID:    toID,
		Content: msg.Content,
	}

	if err := h.svc.SendMessage(r.Context(), message); err != nil {
		h.log.Error().Err(err).Msg("failed to send message")
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func (h *Handler) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	// Get the userID from URL parameters
	userID := chi.URLParam(r, "userID")

	// Get sender ID from context
	fromID, ok := r.Context().Value(auth.KeyUserID).(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the target user ID from URL parameter
	toID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	messages, err := h.svc.GetMessages(r.Context(), fromID, toID, 50) // Default limit of 50
	if err != nil {
		h.log.Error().Err(err).Msg("failed to get messages")
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
