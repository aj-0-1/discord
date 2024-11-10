package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Message struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FromID    uuid.UUID `json:"fromId" db:"from_id"`
	ToID      uuid.UUID `json:"toId" db:"to_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type Service struct {
	db    *sql.DB
	redis *redis.Client
	log   *zerolog.Logger
	hub   *Hub
}

func NewService(db *sql.DB, redis *redis.Client, log *zerolog.Logger) *Service {
	svc := &Service{
		db:    db,
		redis: redis,
		log:   log,
	}

	svc.hub = NewHub(redis, log)
	go svc.hub.Run()

	return svc
}

func (s *Service) SendMessage(ctx context.Context, msg *Message) error {
	msg.CreatedAt = time.Now()
	msg.UpdatedAt = msg.CreatedAt

	const q = `
        INSERT INTO messages (id, from_id, to_id, content, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(ctx, q,
		msg.ID,
		msg.FromID,
		msg.ToID,
		msg.Content,
		msg.CreatedAt,
		msg.UpdatedAt,
	).Scan(&msg.ID, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to store message: %w", err)
	}

	channel := fmt.Sprintf("user:%s:messages", msg.ToID.String())

	if err := s.redis.Publish(ctx, channel, msg).Err(); err != nil {
		s.log.Error().Err(err).
			Str("channel", channel).
			Str("fromId", msg.FromID.String()).
			Str("toId", msg.ToID.String()).
			Msg("failed to publish message to redis")
	} else {
		s.log.Info().
			Str("channel", channel).
			Str("fromId", msg.FromID.String()).
			Str("toId", msg.ToID.String()).
			Msg("message published to redis")
	}

	return nil
}

func (s *Service) GetMessages(ctx context.Context, userID1, userID2 uuid.UUID, limit int) ([]Message, error) {
	const q = `
        SELECT id, from_id, to_id, content, created_at, updated_at
        FROM messages
        WHERE (from_id = $1 AND to_id = $2) OR (from_id = $2 AND to_id = $1)
        ORDER BY created_at DESC
        LIMIT $3`

	rows, err := s.db.QueryContext(ctx, q, userID1, userID2, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(
			&msg.ID,
			&msg.FromID,
			&msg.ToID,
			&msg.Content,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

func (m *Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}
