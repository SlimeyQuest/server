package player

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/entity"
	"github.com/slimeyquest/server/internal/middleware"
	playerSvc "github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/session"
	"github.com/slimeyquest/server/pkg/response"
	"log/slog"
)

const (
	RouteCreateRole         = "/player/role"
	RouteChestOpen          = "/equipment/chests/open"
	RouteDecomposeEquipment = "/equipment/decompose"
	RouteUpgradeChest       = "/equipment/chests/upgrade"
	RouteEquipItem          = "/equipment/equip"
	RouteDrawSkill          = "/skills/draw"
	RouteDrawCompanion      = "/companions/draw"
)

// Handler serves player, equipment, skill and companion endpoints.
type Handler struct {
	loop     *playerSvc.ClosedLoopService
	sessions *session.Manager
	log      *slog.Logger
}

// NewHandler creates a player handler.
func NewHandler(loop *playerSvc.ClosedLoopService, sessions *session.Manager, log *slog.Logger) *Handler {
	return &Handler{loop: loop, sessions: sessions, log: log}
}

// Register mounts player routes on the given router group.
func Register(r gin.IRoutes, h *Handler) {
	r.POST(RouteCreateRole, h.CreateRole)
	r.POST(RouteChestOpen, h.ChestOpen)
	r.POST(RouteDecomposeEquipment, h.DecomposeEquipment)
	r.POST(RouteUpgradeChest, h.UpgradeChest)
	r.POST(RouteEquipItem, h.EquipItem)
	r.POST(RouteDrawSkill, h.DrawSkill)
	r.POST(RouteDrawCompanion, h.DrawCompanion)
}

func (h *Handler) CreateRole(c *gin.Context) {
	auth, ok := middleware.AuthFromContext(c)
	if !ok || !h.sessions.Validate(auth.PlayerID, auth.Token) {
		response.WriteUnauthorized(c)
		return
	}
	var req entity.CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}
	res, err := h.loop.CreateRole(c.Request.Context(), auth.PlayerID, req.DisplayName)
	if err != nil {
		h.log.Error("create_role_failed", "player_id", auth.PlayerID, "error", err)
		response.WriteInternal(c)
		return
	}
	if response.HasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) ChestOpen(c *gin.Context) {
	auth, ok := middleware.AuthFromContext(c)
	if !ok || !h.sessions.Validate(auth.PlayerID, auth.Token) {
		response.WriteUnauthorized(c)
		return
	}
	var req entity.ChestOpenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}
	res, err := h.loop.OpenChest(c.Request.Context(), auth.PlayerID, req.Count)
	if err != nil {
		h.log.Error("chest_open_failed", "player_id", auth.PlayerID, "error", err)
		response.WriteInternal(c)
		return
	}
	if response.HasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) DecomposeEquipment(c *gin.Context) {
	auth, ok := middleware.AuthFromContext(c)
	if !ok || !h.sessions.Validate(auth.PlayerID, auth.Token) {
		response.WriteUnauthorized(c)
		return
	}
	var req entity.DecomposeEquipmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}
	res, err := h.loop.DecomposeEquipment(c.Request.Context(), auth.PlayerID, req.EquipmentUID)
	if err != nil {
		h.log.Error("decompose_failed", "player_id", auth.PlayerID, "error", err)
		response.WriteInternal(c)
		return
	}
	if response.HasAPIError(res.Error) {
		status := http.StatusBadRequest
		if res.Error.Code == entity.ErrorCodeNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) UpgradeChest(c *gin.Context) {
	auth, ok := middleware.AuthFromContext(c)
	if !ok || !h.sessions.Validate(auth.PlayerID, auth.Token) {
		response.WriteUnauthorized(c)
		return
	}
	var req entity.UpgradeChestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}
	res, err := h.loop.UpgradeChest(c.Request.Context(), auth.PlayerID, req.TargetLevel)
	if err != nil {
		h.log.Error("upgrade_chest_failed", "player_id", auth.PlayerID, "error", err)
		response.WriteInternal(c)
		return
	}
	if response.HasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) EquipItem(c *gin.Context) {
	auth, ok := middleware.AuthFromContext(c)
	if !ok || !h.sessions.Validate(auth.PlayerID, auth.Token) {
		response.WriteUnauthorized(c)
		return
	}
	var req entity.EquipItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}
	slot, ok := entity.ParseEquipmentSlot(req.Slot)
	if !ok {
		response.WriteBadRequest(c, "invalid equipment slot")
		return
	}
	res, err := h.loop.EquipItem(c.Request.Context(), auth.PlayerID, req.EquipmentUID, slot)
	if err != nil {
		h.log.Error("equip_failed", "player_id", auth.PlayerID, "error", err)
		response.WriteInternal(c)
		return
	}
	if response.HasAPIError(res.Error) {
		status := http.StatusBadRequest
		if res.Error.Code == entity.ErrorCodeNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) DrawSkill(c *gin.Context) {
	auth, ok := middleware.AuthFromContext(c)
	if !ok || !h.sessions.Validate(auth.PlayerID, auth.Token) {
		response.WriteUnauthorized(c)
		return
	}
	var req entity.DrawSkillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}
	res, err := h.loop.DrawSkill(c.Request.Context(), auth.PlayerID, req.DrawCount)
	if err != nil {
		h.log.Error("draw_skill_failed", "player_id", auth.PlayerID, "error", err)
		response.WriteInternal(c)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) DrawCompanion(c *gin.Context) {
	auth, ok := middleware.AuthFromContext(c)
	if !ok || !h.sessions.Validate(auth.PlayerID, auth.Token) {
		response.WriteUnauthorized(c)
		return
	}
	var req entity.DrawCompanionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}
	res, err := h.loop.DrawCompanion(c.Request.Context(), auth.PlayerID, req.DrawCount)
	if err != nil {
		h.log.Error("draw_companion_failed", "player_id", auth.PlayerID, "error", err)
		response.WriteInternal(c)
		return
	}
	c.JSON(http.StatusOK, res)
}
