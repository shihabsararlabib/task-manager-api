# Task Manager API (Go)

Intermediate-level Golang REST API project using PostgreSQL, Docker, JWT authentication, and layered architecture.

## Features

- User registration and login with JWT access/refresh tokens
- Refresh and logout token flows
- User-scoped CRUD endpoints for tasks
- Pagination and status filtering for task listing
- Role-based admin endpoint (`admin` role)
- Service + repository separation
- PostgreSQL persistence
- SQL migrations folder
- Docker + docker-compose setup
- Unit tests for service layer

## Tech Stack

- Go 1.22
- chi router
- pgx PostgreSQL driver

## Project Structure

- cmd/server: application entrypoint
- internal/config: environment and app config
- internal/database: PostgreSQL connection setup
- internal/models: domain models
- internal/repository: persistence layer
- internal/service: business logic
- internal/handlers: HTTP handlers
- internal/router: route setup
- migrations: SQL migration scripts

## Setup

1. Copy `.env.example` to `.env` (optional).
2. Start PostgreSQL via Docker:
   - `docker compose up -d db`
3. Apply SQL in migration order:
  - `migrations/000001_create_tasks.up.sql`
  - `migrations/000002_users_and_task_owner.up.sql`
  - `migrations/000003_roles_and_refresh_tokens.up.sql`
4. Run:
   - `go mod tidy`
   - `go run ./cmd/server`

API runs at `http://localhost:8080`.

## Frontend (React + Vite)

1. Open `frontend/.env.example` and copy it to `frontend/.env` if needed.
2. Install frontend dependencies:
  - `make frontend-install`
3. Start frontend:
  - `make frontend-dev`

Frontend runs at `http://localhost:5173`.

## API Endpoints

- `GET /health`
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `POST /auth/refresh`
- `POST /auth/logout`
- `POST /tasks` (requires Bearer token)
- `GET /tasks` (requires Bearer token)
- `GET /tasks/{id}` (requires Bearer token)
- `PUT /tasks/{id}` (requires Bearer token)
- `DELETE /tasks/{id}` (requires Bearer token)
- `GET /admin/users` (requires Bearer token with `admin` role)

### Auth Register Payload

```json
{
  "name": "Alice",
  "email": "alice@example.com",
  "password": "password123"
}
```

### Auth Login Payload

```json
{
  "email": "alice@example.com",
  "password": "password123"
}
```

Login/Register response contains a JWT token. Send it as:

`Authorization: Bearer <token>`

`/auth/login` and `/auth/register` now return:
- `access_token` (short-lived, used in `Authorization` header)
- `refresh_token` (long-lived, used with `/auth/refresh`)

### Sample Create Payload

```json
{
  "title": "Write docs",
  "description": "Finish project README"
}
```

### Sample Update Payload

```json
{
  "title": "Write docs",
  "description": "README and API guide",
  "status": "in_progress"
}
```

Allowed status values: `todo`, `in_progress`, `done`.

Task listing supports query params:
- `GET /tasks?page=1&limit=20`
- `GET /tasks?status=todo&page=1&limit=10`

## Commands

- `make run` - run API
- `make build` - build binary
- `make test` - run tests
- `make lint` - run `go vet`
- `make docker-up` - start app and db containers
- `make docker-down` - stop and remove containers
