package main

import (
	"context"
	"payment/config"
	"payment/internal/app"
	"payment/internal/domain/action"
	"payment/pkg/logger"
)

func main() {
	ctx, cfg := context.Background(), config.MustLoad()

	log := logger.New(cfg.DevLevel)
	log.Info(ctx, action.ServiceSetup, "Logger and configuration setup has been finished...")

	a := app.New(ctx, cfg, log)
	log.Info(ctx, action.ServiceSetup, "Application setup has been finished...")

	a.Start(ctx)
	defer a.Stop(ctx)
}
