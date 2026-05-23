package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/apitypes"
)

func (s *Server) handleClaimIdle(c *gin.Context) {
	auth, ok := authFromContext(c)
	if !ok {
		writeUnauthorized(c)
		return
	}
	if !s.sessions.Validate(auth.PlayerID, auth.Token) {
		writeUnauthorized(c)
		return
	}

	var req apitypes.ClaimIdleRewardsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}

	res, err := s.idle.Claim(c.Request.Context(), auth.PlayerID, req.ClaimedThroughMs)
	if err != nil {
		s.log.Error("claim_idle_failed", "player_id", auth.PlayerID, "error", err)
		writeInternal(c)
		return
	}
	c.JSON(http.StatusOK, res)
}
