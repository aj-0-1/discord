package database

import (
	"database/sql"
	"discord/internal/config"
	"fmt"

	_ "github.com/lib/pq"
)

func New(cfg *config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}
