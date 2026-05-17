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
// When Go codegen is enabled from the proto repository (buf generate), add to server go.mod:
//
//	require github.com/slimeyquest/proto v0.0.0
//	replace github.com/slimeyquest/proto => ../proto
//
// Workflow: see ../proto/README.md and server Makefile targets proto-lint / proto-gen.
package protocol
