package routers

import (
	"context"
	paymentv1 "payment/internal/adapters/grpc/payment/v1"
)

type COFServer struct {
	paymentv1.UnimplementedCOFServer
}

func NewCOFServer() *COFServer {
	return &COFServer{}
}

func (s *COFServer) GetCOFbyID(ctx context.Context, req *paymentv1.GetCOFbyIDRequest) (*paymentv1.GetCOFbyIDResponse, error) {
	return &paymentv1.GetCOFbyIDResponse{}, nil
}

func (s *COFServer) GetCOFbyCardNum(context.Context, *paymentv1.GetCOFbyCardRequest) (*paymentv1.GetCOFbyIDResponse, error) {
	return &paymentv1.GetCOFbyIDResponse{}, nil
}

func (s *COFServer) DeactivateCOF(context.Context, *paymentv1.DeactivateCOFRequest) (*paymentv1.DeactivateCOFResponse, error) {
	return &paymentv1.DeactivateCOFResponse{}, nil
}

func (s *COFServer) EnableCOF(context.Context, *paymentv1.EnableCOFRequest) (*paymentv1.EnableCOFResponse, error) {
	return &paymentv1.EnableCOFResponse{}, nil
}
