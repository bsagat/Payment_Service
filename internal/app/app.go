package app

import (
	"context"
	"payment/config"
	grpcserver "payment/internal/adapters/grpc"
	"payment/pkg/logger"
	"payment/pkg/postgres"
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
}

func (a *App) Stop(ctx context.Context) {
	a.log.Info(ctx, "Closing application...")
	a.postgresDB.Pool.Close()
	a.gRPC.Stop()
}
