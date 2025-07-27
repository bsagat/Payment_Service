package grpcserver

import (
	"payment/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func GetOptions(cfg config.GRPCServer) []grpc.ServerOption {
	var opts []grpc.ServerOption

	opts = append(opts,
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSizeMiB*1024*1024),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge:      cfg.MaxConnectionAge,
			MaxConnectionAgeGrace: cfg.MaxConnectionAgeGrace,
		}),
	)
	return opts
}
