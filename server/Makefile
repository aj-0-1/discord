# Database configuration
DB_USER ?= discord_user
DB_PASSWORD ?= discord_password
DB_NAME ?= discord
DB_PORT ?= 5432
REDIS_PORT ?= 6379

# Docker configuration
POSTGRES_CONTAINER = discord_postgres
REDIS_CONTAINER = discord_redis
DOCKER_NETWORK = discord_network

# Database URL for migrations
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: setup start stop restart status logs clean migrate-up migrate-down swagger create-network

# Stop local PostgreSQL (if running)
stop-local-pg:
	@echo "Stopping local PostgreSQL..."
	@sudo service postgresql stop || true

# Clean up existing containers
cleanup:
	@echo "Cleaning up existing containers..."
	@sudo docker rm -f $(POSTGRES_CONTAINER) || true
	@sudo docker rm -f $(REDIS_CONTAINER) || true

# Create Docker network
create-network:
	@echo "Creating Docker network..."
	@sudo docker network create $(DOCKER_NETWORK) || true

# Start services (now with proper cleanup)
start: stop-local-pg cleanup create-network
	@echo "Starting services..."
	# Start PostgreSQL
	@sudo docker run --name $(POSTGRES_CONTAINER) \
		--network $(DOCKER_NETWORK) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		-d postgres:15

	# Start Redis
	@sudo docker run --name $(REDIS_CONTAINER) \
		--network $(DOCKER_NETWORK) \
		-p $(REDIS_PORT):6379 \
		-d redis:7

	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@make migrate-up

setup: create-network
	go mod tidy
	go mod verify

stop:
	sudo docker stop $(POSTGRES_CONTAINER) || true
	sudo docker stop $(REDIS_CONTAINER) || true

restart: stop start

status:
	@echo "=== Service Status ==="
	@sudo docker ps --filter "name=$(POSTGRES_CONTAINER)" --format "PostgreSQL: {{.Status}}"
	@sudo docker ps --filter "name=$(REDIS_CONTAINER)" --format "Redis: {{.Status}}"

logs:
	@echo "=== PostgreSQL Logs ==="
	@sudo docker logs $(POSTGRES_CONTAINER) 2>&1 | tail -n 50
	@echo "\n=== Redis Logs ==="
	@sudo docker logs $(REDIS_CONTAINER) 2>&1 | tail -n 50

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

dev: start run

clean:
	sudo docker stop $(POSTGRES_CONTAINER) || true
	sudo docker rm $(POSTGRES_CONTAINER) || true
	sudo docker stop $(REDIS_CONTAINER) || true
	sudo docker rm $(REDIS_CONTAINER) || true
	sudo docker network rm $(DOCKER_NETWORK) || true

redis-cli:
	sudo docker exec -it $(REDIS_CONTAINER) redis-cli

psql:
	sudo docker exec -it $(POSTGRES_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

GOPATH=$(shell go env GOPATH)
swagger:
	$(GOPATH)/bin/swag init -g cmd/server/main.go
