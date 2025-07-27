package app

import (
	"context"
	"os"
	"os/signal"
	"payment/config"
	grpcserver "payment/internal/adapters/grpc"
	"payment/pkg/logger"
	"payment/pkg/postgres"
	"syscall"
)

type App struct {
	postgresDB *postgres.API
	gRPC       *grpcserver.API
	log        logger.Logger
}

func New(ctx context.Context, cfg config.Config, log logger.Logger) *App {
	gRPCserver := grpcserver.New(ctx, cfg.Server.GRPCServer, log)

	db, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		log.Fatal(ctx, "Failed to connect to the database", "error", err)
	}

	return &App{
		log:        log,
		postgresDB: db,
		gRPC:       gRPCserver,
	}
}

func (a *App) Start(ctx context.Context) {
	a.log.Info(ctx, "Starting application...")

	errCh := make(chan error, 1)
	go a.gRPC.Start(ctx, errCh)

	ListenShutdown(ctx, errCh, a.log)
}

func (a *App) Stop(ctx context.Context) {
	a.log.Info(ctx, "Closing application...")
	a.postgresDB.Pool.Close()
	a.gRPC.Stop()
}

func ListenShutdown(ctx context.Context, errCh chan error, log logger.Logger) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Error(ctx, "Catched error!", "error", err)
		return
	case signal := <-signalCh:
		log.Info(ctx, "Catched shutdown signal!", "signal", signal.String())
		return
	}
}
