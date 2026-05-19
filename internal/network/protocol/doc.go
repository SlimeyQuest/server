// Package protocol defines the WebSocket wire boundary for SlimeyQuest.
//
// Future responsibilities (not implemented in the infrastructure phase):
//   - Length-prefixed protobuf frame encoding and decoding
//   - Opcode-based message routing (request/response/notify)
//   - Optional compression and heartbeat extensions
//
// Generated message types will come from:
//
//	github.com/slimeyquest/proto/gen/go/...
//
// Pin a released proto version in server go.mod:
//
//	require github.com/slimeyquest/proto v0.1.0
//
// Upgrade: go get github.com/slimeyquest/proto@vX.Y.Z
//
// Workflow: see github.com/slimeyquest/proto README and server Makefile targets proto-lint / proto-gen.
package protocol
