# SlimeyQuest Server

Lightweight layered modular monolith backend for SlimeyQuest (HTTP migration in progress + PostgreSQL + Redis).

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

   HTTP API documentation starts at [`docs/api/README.md`](docs/api/README.md).

   Legacy WebSocket endpoint during migration: `ws://localhost:8080/ws`

## Project layout

The server uses a layered modular-monolith layout. Interface, business, and data-access code are stored separately to keep transport concerns out of domain logic and persistence concerns out of interface handlers.

```
cmd/server/                         Entry point
internal/app/                       Application wiring and lifecycle
internal/config/                    Environment configuration
internal/logger/                    Structured logging (slog)
internal/interfaces/network/        Interface layer: legacy WebSocket transport, request routing, connection lifecycle
internal/interfaces/network/protocol/ Legacy protobuf wire boundary while HTTP migration is in progress
internal/services/login/            Business layer: account authentication and transport-agnostic login session flow
internal/services/player/           Business layer: player domain model, progression, equipment, chest logic, repository interface
internal/services/idle/             Business layer: idle reward calculation and claim flow
internal/services/reward/           Business layer: reward application and reward result types
internal/services/stage/            Business layer: stage progression and stage rewards
internal/services/session/          Business layer: in-memory session ownership and validation
internal/data/playerrepo/           Data layer: ent-backed player repository implementation
internal/data/storage/              Data layer: PostgreSQL, Redis, and ent client lifecycle
docs/api/                           HTTP API documentation
```

## API direction

The project is migrating away from protobuf-based client/server integration. New functionality should be exposed through HTTP JSON endpoints documented in [`docs/api/README.md`](docs/api/README.md).

The legacy WebSocket/protobuf adapter remains under `internal/interfaces/network` only for compatibility during migration. Keep new business logic in `internal/services` and call it from interface adapters instead of mixing transport code into services.

## ent dependency

The ent schema and generated ORM code live in the separate `github.com/slimeyquest/ent` module. The server consumes the released module version instead of storing generated ent code directly.
