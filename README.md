# SlimeyQuest Server

Lightweight layered modular monolith backend for SlimeyQuest (HTTP JSON REST + PostgreSQL + Redis).

## Prerequisites

- Go 1.26+
- PostgreSQL (default database: `slimeyquest`)
- Redis

## Local development

1. Copy environment template:

   ```bash
   cp .env.example .env
   ```

2. Ensure PostgreSQL and Redis are running and match `.env` (or defaults in `.env.example`).

3. Build and run:

   ```bash
   make build
   ./bin/server
   ```

   Or:

   ```bash
   go run ./cmd/server
   ```

4. Verify:

   ```bash
   curl http://localhost:8080/health
   ```

   Example response:

   ```json
   {"status":"ok","version":"dev","uptime":"1m2s"}
   ```

   HTTP API documentation: [`../docs/api/http-v1.md`](../docs/api/http-v1.md)（中文契约）

   Architecture: [`../docs/architecture/`](../docs/architecture/) · MVP loop: [`../docs/chest-opener-loop.md`](../docs/chest-opener-loop.md)

## Project layout

```
cmd/server/                         Entry point
cmd/http-smoke/                     HTTP API smoke test client
internal/app/                       Application wiring and lifecycle
internal/entity/                    JSON API request/response types
internal/config/                    Environment + embedded gameplay config
internal/logger/                    Structured logging (slog)
internal/api/                       Gin HTTP routing (domain handlers)
internal/middleware/                Bearer auth middleware
pkg/response/                       Unified JSON error responses
internal/services/login/            Authentication and session flow
internal/services/player/           Player domain, equipment, chest loop
internal/services/idle/             Idle reward calculation and claims
internal/services/reward/             Reward application
internal/services/stage/            Stage progression
internal/services/session/          In-memory session tokens
internal/data/playerrepo/           ent-backed player repository
internal/data/storage/              PostgreSQL, Redis, ent client
internal/config/data/               Embedded gameplay CSV/YAML config
```

## API direction

All client/server gameplay integration uses **HTTP JSON** under `/api/v1`. See [`../docs/api/http-v1.md`](../docs/api/http-v1.md).

Keep business logic in `internal/services` and call it from thin HTTP handlers.

## ent dependency

The ent schema lives in the separate `github.com/slimeyquest/ent` module. Local development uses `replace github.com/slimeyquest/ent => ../ent` in `go.mod`.

## Smoke test

With the server running:

```bash
make http-smoke
```
