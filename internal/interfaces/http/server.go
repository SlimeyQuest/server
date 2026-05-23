package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/services/idle"
	"github.com/slimeyquest/server/internal/services/login"
	"github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/session"
	"github.com/slimeyquest/server/internal/services/stage"
)

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

// Server exposes the HTTP REST API.
type Server struct {
	cfg      *config.Config
	log      *slog.Logger
	login    *login.Service
	idle     *idle.Service
	stage    *stage.Service
	loop     *player.ClosedLoopService
	sessions *session.Manager
	engine   *gin.Engine
	http     *http.Server
	startTime time.Time
}

// NewServer creates the HTTP API server.
func NewServer(
	cfg *config.Config,
	log *slog.Logger,
	loginSvc *login.Service,
	idleSvc *idle.Service,
	stageSvc *stage.Service,
	loopSvc *player.ClosedLoopService,
	sessions *session.Manager,
) *Server {
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{
		cfg:       cfg,
		log:       log,
		login:     loginSvc,
		idle:      idleSvc,
		stage:     stageSvc,
		loop:      loopSvc,
		sessions:  sessions,
		startTime: time.Now(),
	}
	s.engine = gin.New()
	s.engine.Use(gin.Recovery())
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.engine.GET("/health", s.handleHealth)

	v1 := s.engine.Group("/api/v1")
	{
		v1.POST("/auth/guest-login", s.handleGuestLogin)
		v1.POST("/auth/phone-register", s.handlePhoneRegister)
		v1.POST("/auth/phone-login", s.handlePhoneLogin)

		authed := v1.Group("")
		authed.Use(AuthMiddleware(s.sessions))
		{
			authed.POST("/idle/claim", s.handleClaimIdle)
			authed.POST("/stages/push", s.handlePushStage)
			authed.POST("/player/role", s.handleCreateRole)
			authed.POST("/equipment/chests/open", s.handleChestOpen)
			authed.POST("/equipment/decompose", s.handleDecomposeEquipment)
			authed.POST("/equipment/chests/upgrade", s.handleUpgradeChest)
			authed.POST("/equipment/equip", s.handleEquipItem)
			authed.POST("/skills/draw", s.handleDrawSkill)
			authed.POST("/companions/draw", s.handleDrawCompanion)
		}
	}
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{
		Status:  "ok",
		Version: s.cfg.ServerVersion,
		Uptime:  time.Since(s.startTime).Round(time.Second).String(),
	})
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	s.http = &http.Server{
		Addr:              s.cfg.HTTPAddr,
		Handler:           s.engine,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s.http.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}
