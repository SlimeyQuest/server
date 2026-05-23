package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	httpapi "github.com/slimeyquest/server/internal/api"
	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/data/playerrepo"
	"github.com/slimeyquest/server/internal/data/storage"
	"github.com/slimeyquest/server/internal/services/idle"
	"github.com/slimeyquest/server/internal/services/login"
	"github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/reward"
	"github.com/slimeyquest/server/internal/services/session"
	"github.com/slimeyquest/server/internal/services/stage"
)

// App wires infrastructure and runs the server lifecycle.
type App struct {
	cfg      *config.Config
	log      *slog.Logger
	postgres *storage.Postgres
	redis    *storage.Redis
	ent      *storage.Ent
	server   *httpapi.Server
}

// New initializes storage and the HTTP API layer.
func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {
	postgres, err := storage.NewPostgres(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}
	log.Info("postgres connected")

	redis, err := storage.NewRedis(ctx, cfg)
	if err != nil {
		postgres.Close()
		return nil, fmt.Errorf("init redis: %w", err)
	}
	log.Info("redis connected")

	entClient, err := storage.NewEnt(ctx, cfg.PostgresDSN, log)
	if err != nil {
		redis.Close()
		postgres.Close()
		return nil, fmt.Errorf("init ent: %w", err)
	}

	gameplayCfg, err := config.LoadGameplay()
	if err != nil {
		entClient.Close()
		redis.Close()
		postgres.Close()
		return nil, fmt.Errorf("load gameplay config: %w", err)
	}

	playerRepo := playerrepo.New(entClient.Client(), gameplayCfg)
	sessionMgr := session.NewManager()
	rewardSvc := reward.NewService(log.With("component", "reward"), gameplayCfg, playerRepo)
	idleSvc := idle.NewService(log.With("component", "idle"), gameplayCfg, playerRepo, rewardSvc)
	stageSvc := stage.NewService(log.With("component", "stage"), gameplayCfg, playerRepo, rewardSvc)
	loopSvc := player.NewClosedLoopService(playerRepo)
	loginSvc := login.NewService(log.With("component", "login"), playerRepo, sessionMgr, idleSvc, stageSvc)
	server := httpapi.NewServer(cfg, log.With("component", "http"), loginSvc, idleSvc, stageSvc, loopSvc, sessionMgr)

	return &App{
		cfg:      cfg,
		log:      log,
		postgres: postgres,
		redis:    redis,
		ent:      entClient,
		server:   server,
	}, nil
}

// Run starts the server and blocks until ctx is cancelled, then shuts down cleanly.
func (a *App) Run(ctx context.Context) error {
	a.log.Info("starting slimeyquest server",
		"env", a.cfg.AppEnv,
		"addr", a.cfg.HTTPAddr,
	)

	errCh := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		a.log.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := a.shutdown(shutdownCtx); err != nil {
		return err
	}

	a.log.Info("shutdown complete")
	return nil
}

func (a *App) shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Warn("http shutdown error", "error", err)
	}

	a.ent.Close()
	a.postgres.Close()
	if err := a.redis.Close(); err != nil {
		a.log.Warn("redis close error", "error", err)
	}

	return nil
}

// Exit runs the application and exits the process with a non-zero code on failure.
func Exit(log *slog.Logger, err error) {
	if err != nil {
		log.Error("fatal error", "error", err)
		os.Exit(1)
	}
}
