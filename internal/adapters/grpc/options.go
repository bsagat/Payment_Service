package grpcserver

import (
	"payment/config"
	"payment/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func GetOptions(cfg config.GRPCServer, log logger.Logger) []grpc.ServerOption {
	var opts []grpc.ServerOption

	opts = append(opts,
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSizeMiB*1024*1024),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge:      cfg.MaxConnectionAge,
			MaxConnectionAgeGrace: cfg.MaxConnectionAgeGrace,
		}),
		grpc.UnaryInterceptor(LoggingInterceptor(log)),
	)
	return opts
}
