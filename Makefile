APP_NAME=taskmanager

.PHONY: run build test tidy lint migrate-up migrate-down docker-up docker-down frontend-install frontend-dev frontend-build

run:
	go run ./cmd/server

build:
	go build -o bin/$(APP_NAME) ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy

lint:
	go vet ./...

migrate-up:
	@echo "Apply migrations using your migration tool of choice (goose/migrate)."
	@echo "Current migration files are in ./migrations"

migrate-down:
	@echo "Rollback migrations using your migration tool of choice (goose/migrate)."

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build