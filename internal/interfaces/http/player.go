package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/apitypes"
)

func (s *Server) handleCreateRole(c *gin.Context) {
	auth, ok := authFromContext(c)
	if !ok || !s.sessions.Validate(auth.PlayerID, auth.Token) {
		writeUnauthorized(c)
		return
	}
	var req apitypes.CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}
	res, err := s.loop.CreateRole(c.Request.Context(), auth.PlayerID, req.DisplayName)
	if err != nil {
		s.log.Error("create_role_failed", "player_id", auth.PlayerID, "error", err)
		writeInternal(c)
		return
	}
	if hasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) handleChestOpen(c *gin.Context) {
	auth, ok := authFromContext(c)
	if !ok || !s.sessions.Validate(auth.PlayerID, auth.Token) {
		writeUnauthorized(c)
		return
	}
	var req apitypes.ChestOpenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}
	res, err := s.loop.OpenChest(c.Request.Context(), auth.PlayerID, req.Count)
	if err != nil {
		s.log.Error("chest_open_failed", "player_id", auth.PlayerID, "error", err)
		writeInternal(c)
		return
	}
	if hasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) handleDecomposeEquipment(c *gin.Context) {
	auth, ok := authFromContext(c)
	if !ok || !s.sessions.Validate(auth.PlayerID, auth.Token) {
		writeUnauthorized(c)
		return
	}
	var req apitypes.DecomposeEquipmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}
	res, err := s.loop.DecomposeEquipment(c.Request.Context(), auth.PlayerID, req.EquipmentUID)
	if err != nil {
		s.log.Error("decompose_failed", "player_id", auth.PlayerID, "error", err)
		writeInternal(c)
		return
	}
	if hasAPIError(res.Error) {
		status := http.StatusBadRequest
		if res.Error.Code == apitypes.ErrorCodeNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) handleUpgradeChest(c *gin.Context) {
	auth, ok := authFromContext(c)
	if !ok || !s.sessions.Validate(auth.PlayerID, auth.Token) {
		writeUnauthorized(c)
		return
	}
	var req apitypes.UpgradeChestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}
	res, err := s.loop.UpgradeChest(c.Request.Context(), auth.PlayerID, req.TargetLevel)
	if err != nil {
		s.log.Error("upgrade_chest_failed", "player_id", auth.PlayerID, "error", err)
		writeInternal(c)
		return
	}
	if hasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) handleEquipItem(c *gin.Context) {
	auth, ok := authFromContext(c)
	if !ok || !s.sessions.Validate(auth.PlayerID, auth.Token) {
		writeUnauthorized(c)
		return
	}
	var req apitypes.EquipItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}
	slot, ok := apitypes.ParseEquipmentSlot(req.Slot)
	if !ok {
		writeBadRequest(c, "invalid equipment slot")
		return
	}
	res, err := s.loop.EquipItem(c.Request.Context(), auth.PlayerID, req.EquipmentUID, slot)
	if err != nil {
		s.log.Error("equip_failed", "player_id", auth.PlayerID, "error", err)
		writeInternal(c)
		return
	}
	if hasAPIError(res.Error) {
		status := http.StatusBadRequest
		if res.Error.Code == apitypes.ErrorCodeNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) handleDrawSkill(c *gin.Context) {
	auth, ok := authFromContext(c)
	if !ok || !s.sessions.Validate(auth.PlayerID, auth.Token) {
		writeUnauthorized(c)
		return
	}
	var req apitypes.DrawSkillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}
	res, err := s.loop.DrawSkill(c.Request.Context(), auth.PlayerID, req.DrawCount)
	if err != nil {
		s.log.Error("draw_skill_failed", "player_id", auth.PlayerID, "error", err)
		writeInternal(c)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) handleDrawCompanion(c *gin.Context) {
	auth, ok := authFromContext(c)
	if !ok || !s.sessions.Validate(auth.PlayerID, auth.Token) {
		writeUnauthorized(c)
		return
	}
	var req apitypes.DrawCompanionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}
	res, err := s.loop.DrawCompanion(c.Request.Context(), auth.PlayerID, req.DrawCount)
	if err != nil {
		s.log.Error("draw_companion_failed", "player_id", auth.PlayerID, "error", err)
		writeInternal(c)
		return
	}
	c.JSON(http.StatusOK, res)
}
