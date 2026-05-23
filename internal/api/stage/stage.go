package stage

import (
	"net/http"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/entity"
	"github.com/slimeyquest/server/internal/middleware"
	"github.com/slimeyquest/server/internal/services/session"
	stageSvc "github.com/slimeyquest/server/internal/services/stage"
	"github.com/slimeyquest/server/pkg/response"
)

const RoutePush = "/stages/push"

// Handler serves stage progression endpoints.
type Handler struct {
	stage    *stageSvc.Service
	sessions *session.Manager
	log      *slog.Logger
}

// NewHandler creates a stage handler.
func NewHandler(stageService *stageSvc.Service, sessions *session.Manager, log *slog.Logger) *Handler {
	return &Handler{stage: stageService, sessions: sessions, log: log}
}

// Register mounts stage routes on the given router group.
func Register(r gin.IRoutes, h *Handler) {
	r.POST(RoutePush, h.Push)
}

func (h *Handler) Push(c *gin.Context) {
	auth, ok := middleware.AuthFromContext(c)
	if !ok {
		response.WriteUnauthorized(c)
		return
	}
	if !h.sessions.Validate(auth.PlayerID, auth.Token) {
		response.WriteUnauthorized(c)
		return
	}

	var req entity.PushStageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}

	res, err := h.stage.PushStage(c.Request.Context(), auth.PlayerID, req.TargetStageIndex)
	if err != nil {
		h.log.Error("push_stage_failed", "player_id", auth.PlayerID, "error", err)
		response.WriteInternal(c)
		return
	}
	c.JSON(http.StatusOK, res)
}
