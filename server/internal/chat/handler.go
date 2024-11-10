package chat

import (
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
	r.Get("/messages/{userID}", h.handleGetMessages)
	r.Get("/ws", h.handleWebSocket)

	return r
}

// @Summary WebSocket connection
// @Description Connect to WebSocket for real-time messages
// @Tags chat
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 101 {string} string "Switching protocols"
// @Failure 401 {string} string "Unauthorized"
// @Router /chat/ws [get]
func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		h.log.Error().Msg("user ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.log.Debug().
		Str("userId", userID.String()).
		Str("remoteAddr", r.RemoteAddr).
		Msg("websocket connection attempt")

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

// @Summary Send message
// @Description Send a private message to another user
// @Tags chat
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body Message true "Message content"
// @Success 200 {object} Message
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Unauthorized"
// @Router /chat/messages [post]
func (h *Handler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var msg struct {
		ToID    string `json:"toId"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fromID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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

// @Summary Get messages
// @Description Get chat messages with another user
// @Tags chat
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param userID path string true "User ID to get messages with"
// @Success 200 {array} Message
// @Failure 401 {string} string "Unauthorized"
// @Router /chat/messages/{userID} [get]
func (h *Handler) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	fromID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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
