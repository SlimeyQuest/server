package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/slimeyquest/server/internal/app"
	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg)
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg, log)
	if err != nil {
		app.Exit(log, err)
	}

	app.Exit(log, application.Run(ctx))
}
