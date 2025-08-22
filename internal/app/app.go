package app

import (
	"context"
	"os"
	"os/signal"
	"payment/config"
	"payment/internal/adapters/broker/bereke"
	grpcserver "payment/internal/adapters/grpc"
	"payment/internal/adapters/repo"
	"payment/internal/domain/action"
	"payment/internal/service"
	"payment/pkg/logger"
	"payment/pkg/postgres"
	"syscall"

	"github.com/bsagat/bereke-merchant-api/models/types"
)

type App struct {
	postgresDB *postgres.API
	gRPC       *grpcserver.API
	log        logger.Logger
}

func New(ctx context.Context, cfg config.Config, log logger.Logger) *App {
	db, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		log.Fatal(ctx, action.ServiceStartFail, err, "Failed to connect to the database")
	}
	log.Info(ctx, action.DbConnected, "Database connection has been estabilished")

	brokerMerchant, err := bereke.NewClient(cfg.Broker.Login, cfg.Broker.Password, types.Mode(cfg.Broker.Mode))
	if err != nil {
		log.Fatal(ctx, action.ServiceStartFail, err, "Failed to create broker merchant")
	}
	log.Info(ctx, action.DbConnected, "Merchant broker has been created")

	paymentRepo := repo.NewPostgresPaymentRepo(db.Pool)
	paymentService := service.NewPaymentService(brokerMerchant, paymentRepo, log)
	gRPCserver := grpcserver.New(ctx, cfg.Server.GRPCServer, paymentService, log)

	return &App{
		log:        log,
		postgresDB: db,
		gRPC:       gRPCserver,
	}
}

func (a *App) Start(ctx context.Context) {
	a.log.Info(ctx, action.ServiceStarted, "Starting application...")

	errCh := make(chan error, 1)
	go a.gRPC.Start(ctx, errCh)

	ListenShutdown(ctx, errCh, a.log)
}

func (a *App) Stop(ctx context.Context) {
	a.log.Info(ctx, action.GracefulShutdown, "Closing application...")
	a.postgresDB.Pool.Close()
	a.gRPC.Stop()
	a.log.Info(ctx, action.GracefulShutdown, "Application has been closed...")
}

func ListenShutdown(ctx context.Context, errCh chan error, log logger.Logger) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Error(ctx, action.GracefulShutdown, err, "Catched error!")
		return
	case signal := <-signalCh:
		log.Info(ctx, action.GracefulShutdown, "Catched shutdown signal!", "signal", signal.String())
		return
	}
}
