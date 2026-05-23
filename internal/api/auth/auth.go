package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/entity"
	"github.com/slimeyquest/server/internal/services/login"
	"github.com/slimeyquest/server/pkg/response"
)

const (
	RouteGuestLogin    = "/auth/guest-login"
	RoutePhoneRegister = "/auth/phone-register"
	RoutePhoneLogin    = "/auth/phone-login"
)

// Handler serves authentication endpoints.
type Handler struct {
	login *login.Service
}

// NewHandler creates an auth handler.
func NewHandler(loginSvc *login.Service) *Handler {
	return &Handler{login: loginSvc}
}

// Register mounts auth routes on the given router group.
func Register(r gin.IRoutes, h *Handler) {
	r.POST(RouteGuestLogin, h.GuestLogin)
	r.POST(RoutePhoneRegister, h.PhoneRegister)
	r.POST(RoutePhoneLogin, h.PhoneLogin)
}

func (h *Handler) GuestLogin(c *gin.Context) {
	var req entity.GuestLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}
	res, _ := h.login.GuestLogin(c.Request.Context(), &req)
	if response.HasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) PhoneRegister(c *gin.Context) {
	var req entity.PhoneRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}
	res, _ := h.login.PhoneRegister(c.Request.Context(), &req)
	if response.HasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) PhoneLogin(c *gin.Context) {
	var req entity.PhoneLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteBadRequest(c, "invalid json body")
		return
	}
	res, _ := h.login.PhoneLogin(c.Request.Context(), &req)
	if response.HasAPIError(res.Error) {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}
