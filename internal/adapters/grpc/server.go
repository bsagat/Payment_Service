package grpcserver

import (
	"context"
	"fmt"
	"net"
	"payment/config"
	paymentv1 "payment/internal/adapters/grpc/payment/v1"
	"payment/internal/adapters/grpc/routers"
	"payment/internal/domain/action"
	"payment/internal/domain/ports"
	"payment/pkg/logger"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type API struct {
	server *grpc.Server
	cfg    config.GRPCServer

	log logger.Logger
}

func New(ctx context.Context, cfg config.GRPCServer, paymentService ports.PaymentService, log logger.Logger) *API {
	mux := runtime.NewServeMux()
	server := grpc.NewServer(GetOptions(cfg, log)...)

	paymentv1.RegisterPaymentServer(server, routers.NewPaymentServer(paymentService, log))
	paymentv1.RegisterPaymentHandlerFromEndpoint(ctx, mux, "localhost:7000", nil)

	return &API{
		server: server,
		cfg:    cfg,
		log:    log,
	}
}

func (a *API) Start(ctx context.Context, errCh chan error) {
	l, err := net.Listen("tcp", ":"+a.cfg.Port)
	if err != nil {
		a.log.Error(ctx, action.ServerStartFail, err, "Failed to listen on port", "port", a.cfg.Port)
		errCh <- fmt.Errorf("failed to listen on port %s: %w", a.cfg.Port, err)
		return
	}

	a.log.Info(ctx, action.ServerStarted, "Server has been started", "port", a.cfg.Port)
	if err := a.server.Serve(l); err != nil {
		a.log.Error(ctx, action.ServerStartFail, err, "Failed to start gRPC server")
		errCh <- fmt.Errorf("failed to start gRPC server: %w", err)
		return
	}

	a.log.Info(ctx, action.ServerClosed, "Server has been stopped")
}

func (a *API) Stop() {
	a.server.GracefulStop()
}
