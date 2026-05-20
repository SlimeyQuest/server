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

Definitions live in the sibling `../proto` repository. Generated Go packages import as:

```go
import gatewayv1 "github.com/slimeyquest/proto/gen/go/gateway"
```

`go.mod` depends on the released `github.com/slimeyquest/proto` module. Do not commit a local `replace` for proto in this repository; use a Go workspace locally if you need to develop server and proto side by side.

For private module downloads in CI, configure `GOPRIVATE=github.com/slimeyquest/*` and provide a GitHub token with read access to the proto repository.

Regenerate after proto changes in the proto repository:

```bash
make proto-gen    # delegates to ../proto Makefile (Go + TS)
go build ./...
```

Full client sync: `../tools/scripts/sync-proto.sh`.

See [internal/network/protocol/doc.go](internal/network/protocol/doc.go) for the wire-layer boundary.

## Deferred

- **ent ORM** — introduced later with player schema, login persistence, and gameplay data modeling (not part of the infrastructure foundation).
