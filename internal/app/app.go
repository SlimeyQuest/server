package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/idle"
	"github.com/slimeyquest/server/internal/login"
	"github.com/slimeyquest/server/internal/network"
	"github.com/slimeyquest/server/internal/player"
	"github.com/slimeyquest/server/internal/reward"
	"github.com/slimeyquest/server/internal/session"
	"github.com/slimeyquest/server/internal/stage"
	"github.com/slimeyquest/server/internal/storage"
)

// App wires infrastructure and runs the server lifecycle.
type App struct {
	cfg      *config.Config
	log      *slog.Logger
	postgres *storage.Postgres
	redis    *storage.Redis
	ent      *storage.Ent
	hub      *network.Hub
	server   *network.Server
}

// New initializes storage and the network layer.
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

	gameplayCfg, err := gameplayconfig.Load()
	if err != nil {
		entClient.Close()
		redis.Close()
		postgres.Close()
		return nil, fmt.Errorf("load gameplay config: %w", err)
	}

	playerRepo := player.NewRepository(entClient.Client(), gameplayCfg)
	sessionMgr := session.NewManager()
	rewardSvc := reward.NewService(log.With("component", "reward"), gameplayCfg, playerRepo)
	idleSvc := idle.NewService(log.With("component", "idle"), gameplayCfg, playerRepo, rewardSvc)
	stageSvc := stage.NewService(log.With("component", "stage"), gameplayCfg, playerRepo, rewardSvc)
	loginSvc := login.NewService(log.With("component", "login"), playerRepo, sessionMgr, idleSvc, stageSvc)
	gameplayHandler := network.NewGameplay(idleSvc, stageSvc, sessionMgr)
	hub := network.NewHub(log.With("component", "hub"), sessionMgr)
	server := network.NewServer(cfg, log.With("component", "http"), hub, loginSvc, gameplayHandler)

	return &App{
		cfg:      cfg,
		log:      log,
		postgres: postgres,
		redis:    redis,
		ent:      entClient,
		hub:      hub,
		server:   server,
	}, nil
}

// Run starts the server and blocks until ctx is cancelled, then shuts down cleanly.
func (a *App) Run(ctx context.Context) error {
	a.log.Info("starting slimeyquest server",
		"env", a.cfg.AppEnv,
		"addr", a.cfg.HTTPAddr,
	)

	appCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.hub.Run(appCtx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	select {
	case <-ctx.Done():
		a.log.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			a.shutdown(appCtx)
			wg.Wait()
			return fmt.Errorf("http server: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := a.shutdown(shutdownCtx); err != nil {
		return err
	}

	cancel()
	wg.Wait()
	a.log.Info("shutdown complete")
	return nil
}

func (a *App) shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Warn("http shutdown error", "error", err)
	}

	a.hub.CloseAll()
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
