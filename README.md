# PodVibe

Backend MVP for PodVibe — social network for bite-sized podcasts.

## Stack
- Go 1.21
- Gin, GORM (PostgreSQL), JWT, zap
- Docker + docker-compose (Postgres + Redis)

## Quick start
1. Create `.env` (not committed; `.gitignore` added) from `.env.example` and set secrets:
   - `DB_DSN` (required)
   - `JWT_SECRET` (required)
   - optional: `APP_PORT`, `REDIS_ADDR`, `STORAGE_PATH`, `ACCESS_TOKEN_MINUTES`, `REFRESH_TOKEN_DAYS`
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
- Database schema is created via GORM auto-migrations on startup (or apply `migrations/001_init.sql` manually).
- Refresh tokens are persisted in `refresh_tokens` table; access tokens are JWTs.
- File uploads are stored under `storage/` and exposed via `/static/...`.
- Transcription is a dummy stub and marks episodes as `ready` with placeholder text.

## API examples (curl)
- Register:
  ```bash
  curl -X POST http://localhost:8080/auth/register \
    -H "Content-Type: application/json" \
    -d '{"email":"user@example.com","username":"user123","password":"secret123","display_name":"User"}'
  ```
- Login:
  ```bash
  curl -X POST http://localhost:8080/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"user@example.com","password":"secret123"}'
  ```
- Create podcast (with cover):
  ```bash
  curl -X POST http://localhost:8080/podcasts \
    -H "Authorization: Bearer <ACCESS_TOKEN>" \
    -F "title=My Podcast" \
    -F "description=About stuff" \
    -F "cover=@/path/to/cover.jpg"
  ```
- Create episode (audio <3 minutes, mp3/wav):
  ```bash
  curl -X POST http://localhost:8080/podcasts/1/episodes \
    -H "Authorization: Bearer <ACCESS_TOKEN>" \
    -F "title=Episode 1" \
    -F "description=Quick intro" \
    -F "tags=#news,#tech" \
    -F "audio=@/path/to/audio.mp3"
  ```
- Like / comment:
  ```bash
  curl -X POST http://localhost:8080/episodes/1/like -H "Authorization: Bearer <ACCESS_TOKEN>"
  curl -X POST http://localhost:8080/episodes/1/comments \
    -H "Authorization: Bearer <ACCESS_TOKEN>" \
    -H "Content-Type: application/json" \
    -d '{"text":"Great!"}'
  ```
- Feed and popular:
  ```bash
  curl -H "Authorization: Bearer <ACCESS_TOKEN>" "http://localhost:8080/feed?page=1&page_size=20"
  curl "http://localhost:8080/episodes/popular?page=1&page_size=20"
  ```
