package idle

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/entity"
	"github.com/slimeyquest/server/internal/middleware"
	"github.com/slimeyquest/server/internal/services/idle"
	"github.com/slimeyquest/server/internal/services/session"
	"github.com/slimeyquest/server/pkg/response"
	"log/slog"
)

const RouteClaim = "/idle/claim"

// Handler serves idle gameplay endpoints.
type Handler struct {
	idle     *idle.Service
	sessions *session.Manager
	log      *slog.Logger
}

// NewHandler creates an idle handler.
func NewHandler(idleSvc *idle.Service, sessions *session.Manager, log *slog.Logger) *Handler {
	return &Handler{idle: idleSvc, sessions: sessions, log: log}
}

// Register mounts idle routes on the given router group.
func Register(r gin.IRoutes, h *Handler) {
	r.POST(RouteClaim, h.Claim)
}

func (h *Handler) Claim(c *gin.Context) {
	auth, ok := middleware.AuthFromContext(c)
	if !ok {
		response.WriteUnauthorized(c)
		return
	}
	if !h.sessions.Validate(auth.PlayerID, auth.Token) {
		response.WriteUnauthorized(c)
		return
	}

	var req entity.ClaimIdleRewardsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}

	res, err := h.idle.Claim(c.Request.Context(), auth.PlayerID, req.ClaimedThroughMs)
	if err != nil {
		h.log.Error("claim_idle_failed", "player_id", auth.PlayerID, "error", err)
		response.WriteInternal(c)
		return
	}
	c.JSON(http.StatusOK, res)
}
