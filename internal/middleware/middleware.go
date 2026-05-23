package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/services/session"
	"github.com/slimeyquest/server/pkg/response"
)

// AuthContext holds the authenticated player session.
type AuthContext struct {
	PlayerID int64
	Token    string
}

// AuthMiddleware validates bearer tokens for gameplay routes.
func AuthMiddleware(sessions *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			response.WriteUnauthorized(c)
			c.Abort()
			return
		}
		sess, ok := sessions.GetByToken(token)
		if !ok {
			response.WriteUnauthorized(c)
			c.Abort()
			return
		}
		c.Set("auth", AuthContext{PlayerID: sess.PlayerID, Token: sess.Token})
		c.Next()
	}
}

// AuthFromContext returns the authenticated session attached by AuthMiddleware.
func AuthFromContext(c *gin.Context) (AuthContext, bool) {
	v, ok := c.Get("auth")
	if !ok {
		return AuthContext{}, false
	}
	auth, ok := v.(AuthContext)
	return auth, ok
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
