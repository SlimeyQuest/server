package session

import "fmt"

// SessionKey returns the Redis key for a session token.
// Reserved for future Redis-backed session storage.
func SessionKey(token string) string {
	return fmt.Sprintf("session:%s", token)
}
