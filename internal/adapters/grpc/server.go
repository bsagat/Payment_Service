package grpcserver

import (
	"context"
	"fmt"
	"net"
	"payment/config"
	paymentv1 "payment/internal/adapters/grpc/payment/v1"
	"payment/internal/adapters/grpc/routers"
	"payment/pkg/logger"

	"google.golang.org/grpc"
)

type API struct {
	server *grpc.Server
	cfg    config.GRPCServer

	log logger.Logger
}

func New(ctx context.Context, cfg config.GRPCServer, log logger.Logger) *API {
	server := grpc.NewServer(GetOptions(cfg)...)

	paymentv1.RegisterCOFServer(server, routers.NewCOFServer())
	paymentv1.RegisterPaymentServer(server, routers.NewPaymentServer())

	return &API{
		server: server,
		cfg:    cfg,
		log:    log,
	}
}

func (a *API) Start(ctx context.Context, errCh chan error) {
	l, err := net.Listen("tcp", ":"+a.cfg.Port)
	if err != nil {
		a.log.Error(ctx, "Failed to listen on port", "port", a.cfg.Port, "error", err)
		errCh <- fmt.Errorf("failed to listen on port %s: %w", a.cfg.Port, err)
		return
	}

	if err := a.server.Serve(l); err != nil {
		a.log.Error(ctx, "Failed to start gRPC server", "error", err)
		errCh <- fmt.Errorf("failed to start gRPC server: %w", err)
		return
	}
}

func (a *API) Stop() {
	a.server.GracefulStop()
}
