package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/api/auth"
	idleapi "github.com/slimeyquest/server/internal/api/idle"
	"github.com/slimeyquest/server/internal/api/player"
	"github.com/slimeyquest/server/internal/api/stage"
	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/middleware"
	idleSvc "github.com/slimeyquest/server/internal/services/idle"
	"github.com/slimeyquest/server/internal/services/login"
	playerSvc "github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/session"
	stageSvc "github.com/slimeyquest/server/internal/services/stage"
)

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

// Server exposes the HTTP REST API.
type Server struct {
	cfg       *config.Config
	log       *slog.Logger
	login     *login.Service
	idle      *idleSvc.Service
	stage     *stageSvc.Service
	loop      *playerSvc.ClosedLoopService
	sessions  *session.Manager
	engine    *gin.Engine
	http      *http.Server
	startTime time.Time
}

// NewServer creates the HTTP API server.
func NewServer(
	cfg *config.Config,
	log *slog.Logger,
	loginSvc *login.Service,
	idleService *idleSvc.Service,
	stageService *stageSvc.Service,
	loopSvc *playerSvc.ClosedLoopService,
	sessions *session.Manager,
) *Server {
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{
		cfg:       cfg,
		log:       log,
		login:     loginSvc,
		idle:      idleService,
		stage:     stageService,
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
	auth.Register(v1, auth.NewHandler(s.login))

	authed := v1.Group("")
	authed.Use(middleware.AuthMiddleware(s.sessions))
	idleapi.Register(authed, idleapi.NewHandler(s.idle, s.sessions, s.log))
	stage.Register(authed, stage.NewHandler(s.stage, s.sessions, s.log))
	player.Register(authed, player.NewHandler(s.loop, s.sessions, s.log))
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
