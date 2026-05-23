package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/apitypes"
)

func (s *Server) handleGuestLogin(c *gin.Context) {
	var req apitypes.GuestLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}
	res, _ := s.login.GuestLogin(c.Request.Context(), &req)
	if hasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) handlePhoneRegister(c *gin.Context) {
	var req apitypes.PhoneRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}
	res, _ := s.login.PhoneRegister(c.Request.Context(), &req)
	if hasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) handlePhoneLogin(c *gin.Context) {
	var req apitypes.PhoneLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid json body")
		return
	}
	res, _ := s.login.PhoneLogin(c.Request.Context(), &req)
	if hasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}
