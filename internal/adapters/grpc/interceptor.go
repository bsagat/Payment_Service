package grpcserver

import (
	"context"
	"payment/pkg/logger"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Вызов основного обработчика
		resp, err := handler(ctx, req)

		// После вызова — логируем
		st := status.Convert(err)
		log.Info(ctx,
			"grpc_request",
			"handled gRPC request",
			"method", info.FullMethod,
			"duration_ms", time.Since(start).Milliseconds(),
			"error", st.Message(),
			"code", st.Code().String(),
		)

		return resp, err
	}
}
