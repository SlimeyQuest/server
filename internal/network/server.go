package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"

	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/login"
)

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

// Server exposes HTTP health checks and the WebSocket endpoint.
type Server struct {
	cfg       *config.Config
	log       *slog.Logger
	hub       *Hub
	loginSvc  *login.Service
	gameplay  *Gameplay
	http      *http.Server
	connSeq   atomic.Uint64
	startTime time.Time
}

// NewServer creates the HTTP/WebSocket server.
func NewServer(cfg *config.Config, log *slog.Logger, hub *Hub, loginSvc *login.Service, gameplay *Gameplay) *Server {
	return &Server{
		cfg:       cfg,
		log:       log,
		hub:       hub,
		loginSvc:  loginSvc,
		gameplay:  gameplay,
		startTime: time.Now(),
	}
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ws", s.handleWebSocket)

	s.http = &http.Server{
		Addr:              s.cfg.HTTPAddr,
		Handler:           mux,
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

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:  "ok",
		Version: s.cfg.ServerVersion,
		Uptime:  time.Since(s.startTime).Round(time.Second).String(),
	})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.IsDevelopment() && r.Header.Get("Origin") == "" {
		http.Error(w, "origin required", http.StatusForbidden)
		return
	}

	opts := &websocket.AcceptOptions{}
	if s.cfg.IsDevelopment() {
		opts.OriginPatterns = []string{"*"}
	}

	ws, err := websocket.Accept(w, r, opts)
	if err != nil {
		s.log.Warn("websocket accept failed", "error", err)
		return
	}

	seq := s.connSeq.Add(1)
	connID := fmt.Sprintf("conn-%d", seq)
	conn := newConn(connID, s.log, ws, s.hub, s.loginSvc, s.gameplay)
	conn.Serve()
}
