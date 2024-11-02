# Database configuration
DB_USER ?= discord_user
DB_PASSWORD ?= discord_password
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= discord

DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: setup init-db migrate-up migrate-down

setup:
	go mod tidy
	go mod verify

init-db:
	@echo "Creating database and user..."
	@sudo -u postgres psql -c "CREATE DATABASE $(DB_NAME);" || true
	@sudo -u postgres psql -c "CREATE USER $(DB_USER) WITH ENCRYPTED PASSWORD '$(DB_PASSWORD)';" || true
	@sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $(DB_NAME) TO $(DB_USER);" || true
	@echo "Database and user created successfully"

migrate-up:
	@echo "Running migrations..."
	migrate -path migrations -database "$(DB_URL)" up
	@echo "Migrations completed successfully"

migrate-down:
	@echo "Rolling back migrations..."
	migrate -path migrations -database "$(DB_URL)" down
	@echo "Rollback completed successfully"

run:
	go run cmd/server/main.go
