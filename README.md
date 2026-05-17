# SlimeyQuest Server

Lightweight modular monolith backend for SlimeyQuest (WebSocket + PostgreSQL + Redis).

## Prerequisites

- Go 1.26+
- PostgreSQL (default database: `slimeyquest`)
- Redis
- [buf](https://buf.build/docs/installation) (optional, for protobuf codegen)

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

   WebSocket endpoint: `ws://localhost:8080/ws`

## Project layout

```
cmd/server/                 Entry point
internal/app/               Application wiring and lifecycle
internal/config/            Environment configuration
internal/logger/            Structured logging (slog)
internal/network/           HTTP, WebSocket hub, connection lifecycle
internal/network/protocol/  Protobuf wire boundary (future framing/routing)
internal/storage/           PostgreSQL and Redis clients
pkg/                        Reserved for future shared public packages (empty)
```

## Protobuf

Definitions live in the sibling `../proto` repository. Generated Go code will target `github.com/slimeyquest/proto/gen/go/...`.

```bash
make proto-lint   # lint protos
make proto-gen    # generate Go (requires buf.gen.yaml in ../proto)
```

When generated code exists, add to `go.mod`:

```
require github.com/slimeyquest/proto v0.0.0
replace github.com/slimeyquest/proto => ../proto
```

See [internal/network/protocol/doc.go](internal/network/protocol/doc.go) for the wire-layer boundary.

## Deferred

- **ent ORM** — introduced later with player schema, login persistence, and gameplay data modeling (not part of the infrastructure foundation).
