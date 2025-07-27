package main

import (
	"context"
	"payment/config"
	"payment/internal/app"
	"payment/pkg/logger"
)

func main() {
	ctx, cfg := context.Background(), config.MustLoad()

	log := logger.New(cfg.DevLevel)
	log.Info(ctx, "Logger and configuration setup has been finished...")

	a := app.New(ctx, cfg, log)
	log.Info(ctx, "Application setup has been finished...")

	a.Start(ctx)
	defer a.Stop(ctx)
}
