# PodVibe

Backend MVP for PodVibe — social network for bite-sized podcasts.

## Stack
- Go 1.21
- Gin, GORM (PostgreSQL), JWT, zap
- Docker + docker-compose (Postgres + Redis)

## Quick start
1. Copy `.env.example` (or set env vars) with at least:
   - `APP_PORT=8080`
   - `DB_DSN=postgres://postgres:postgres@localhost:5432/podvibe?sslmode=disable`
   - `JWT_SECRET=dev-secret`
   - `STORAGE_PATH=./storage`
2. Run locally:
   ```bash
   go run ./cmd/api
   ```
   or with Docker:
   ```bash
   docker-compose up --build
   ```
3. API available at `http://localhost:8080`, static files served from `/static`.

## Project layout
```
cmd/api            # entrypoint
internal/config    # env config loader
internal/server    # router setup
internal/handlers  # HTTP handlers
internal/services  # business logic
internal/repositories # DB access
internal/models    # GORM models
internal/storage   # local file storage
internal/auth      # JWT helpers & middleware
internal/transcript# dummy transcription service
```

## Notes
- Database schema is created via GORM auto-migrations on startup.
- Refresh tokens are persisted in `refresh_tokens` table; access tokens are JWTs.
- File uploads are stored under `storage/` and exposed via `/static/...`.
- Transcription is a dummy stub and marks episodes as `ready` with placeholder text.
