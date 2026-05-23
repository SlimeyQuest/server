# SlimeyQuest HTTP API

This document is the source of truth for the server HTTP API while the project migrates away from protobuf-based client/server integration.

## API conventions

- Base path: `/api/v1`
- Request and response format: JSON
- Authentication: bearer token returned by login endpoints
- Time values: Unix milliseconds for client-facing timestamps unless otherwise noted
- Error format:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "human readable error"
  }
}
```

## Health

### `GET /health`

Returns process health and uptime.

Response:

```json
{
  "status": "ok",
  "version": "dev",
  "uptime": "1m2s"
}
```

## Authentication

### `POST /api/v1/auth/guest-login`

Creates or resumes a guest player by device id.

Request:

```json
{
  "deviceId": "device-a",
  "clientVersion": "1.0.0"
}
```

Response:

```json
{
  "sessionToken": "token",
  "playerId": 1,
  "profile": {},
  "idleState": {},
  "stageState": {}
}
```

### `POST /api/v1/auth/phone-register`

Creates or resumes a phone account with the MVP verification-code flow.

Request:

```json
{
  "phone": "13800000000",
  "verifyCode": "000000",
  "clientVersion": "1.0.0"
}
```

### `POST /api/v1/auth/phone-login`

Logs in an existing or test-created phone account.

Request:

```json
{
  "phone": "13800000000",
  "verifyCode": "123456",
  "clientVersion": "1.0.0"
}
```

## Gameplay

All gameplay endpoints require:

```http
Authorization: Bearer <sessionToken>
```

### `POST /api/v1/idle/claim`

Claims idle rewards.

Request:

```json
{
  "claimedThroughMs": 1760000000000
}
```

### `POST /api/v1/stages/push`

Attempts to clear the current stage target.

Request:

```json
{
  "targetStageIndex": 1
}
```

### `POST /api/v1/player/role`

Creates or updates the current player role display name.

Request:

```json
{
  "displayName": "SlimeHero"
}
```

### `POST /api/v1/equipment/chests/open`

Opens one or more chests.

Request:

```json
{
  "count": 1
}
```

### `POST /api/v1/equipment/decompose`

Decomposes one unequipped equipment item.

Request:

```json
{
  "equipmentUid": 123
}
```

### `POST /api/v1/equipment/chests/upgrade`

Upgrades the chest opener level.

Request:

```json
{
  "targetLevel": 2
}
```

### `POST /api/v1/equipment/equip`

Equips an owned item.

Request:

```json
{
  "equipmentUid": 123,
  "slot": "WEAPON"
}
```

### `POST /api/v1/skills/draw`

Runs MVP skill shop draws.

Request:

```json
{
  "drawCount": 1
}
```

### `POST /api/v1/companions/draw`

Runs MVP companion shop draws.

Request:

```json
{
  "drawCount": 1
}
```

## Migration notes

The current runtime still contains WebSocket/protobuf adapters under `internal/interfaces/network` for compatibility. New client/server integration should target the HTTP API surface above. The service layer is transport-agnostic and can be called from both the legacy adapter and future HTTP handlers.
