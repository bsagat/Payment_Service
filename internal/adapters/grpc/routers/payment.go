package routers

import (
	"context"
	"errors"
	paymentv1 "payment/internal/adapters/grpc/payment/v1"
	"payment/internal/domain/ports"
	"payment/internal/service"
	"payment/pkg/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentServer struct {
	service ports.PaymentService
	log     logger.Logger
	paymentv1.UnimplementedPaymentServer
}

func NewPaymentServer(service ports.PaymentService, log logger.Logger) *PaymentServer {
	return &PaymentServer{
		service: service,
		log:     log,
	}
}

func (s *PaymentServer) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	if err := ValidateCreateOrderReq(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request body is invalid: %v", err)
	}

	payment, paymentUrl, err := s.service.CreatePayment(ctx, req.OrderId, req.UserId, req.Amount, req.Currency, req.Operation, req.ReturnUrl, req.ErrorUrl)
	if err != nil {
		return nil, status.Errorf(GetGrpcCode(err), "failed to create payment: %v", err)
	}

	return &paymentv1.CreatePaymentResponse{
		PaymentId:  payment.ID,
		PaymentUrl: paymentUrl,
	}, nil
}

func (s *PaymentServer) AuthPayment(ctx context.Context, req *paymentv1.AuthPaymentRequest) (*paymentv1.AuthPaymentResponse, error) {
	if err := ValidateAuthOrderReq(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request body is invalid: %v", err)
	}

	payment, paymentUrl, err := s.service.AuthPayment(ctx, req.OrderId, req.UserId, req.Amount, req.Currency, req.ReturnUrl, req.ErrorUrl)
	if err != nil {
		return nil, status.Errorf(GetGrpcCode(err), "failed to auth payment: %v", err)
	}

	return &paymentv1.AuthPaymentResponse{
		PaymentId:  payment.ID,
		PaymentUrl: paymentUrl,
	}, nil
}

func (s *PaymentServer) DepositPayment(ctx context.Context, req *paymentv1.DepositPaymentRequest) (*paymentv1.DepositPaymentResponse, error) {
	if err := ValidateDepositOrderReq(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request body is invalid: %v", err)
	}

	state, err := s.service.DepositPayment(ctx, req.PaymentId, req.Amount, req.Currency)
	if err != nil {
		return nil, status.Errorf(GetGrpcCode(err), "failed to deposit payment: %v", err)
	}

	return &paymentv1.DepositPaymentResponse{
		Status: string(state),
	}, nil
}

func (s *PaymentServer) GetPayment(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	if err := ValidatePaymentID(req.GetPaymentId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request body is invalid: %v", err)
	}

	payment, err := s.service.GetPayment(ctx, req.PaymentId)
	if err != nil {
		return nil, status.Errorf(GetGrpcCode(err), "failed to get payment: %v", err)
	}

	return &paymentv1.GetPaymentResponse{
		PaymentId: payment.ID,
		OrderId:   payment.OrderID,
		UserId:    payment.UserID,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Status:    string(payment.Status),
		CreatedAt: timestamppb.New(payment.CreatedAt),
		Operation: string(payment.Operation),
		Broker:    payment.Broker,
	}, nil
}

func (s *PaymentServer) GetPaymentStatus(ctx context.Context, req *paymentv1.GetPaymentStatusRequest) (*paymentv1.GetPaymentStatusResponse, error) {
	if err := ValidatePaymentID(req.GetPaymentId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request body is invalid: %v", err)
	}

	paymentStatus, err := s.service.GetPaymentStatus(ctx, req.PaymentId)
	if err != nil {
		return nil, status.Errorf(GetGrpcCode(err), "failed to get payment status: %v", err)
	}

	return &paymentv1.GetPaymentStatusResponse{
		Status: string(paymentStatus),
	}, nil
}

func (s *PaymentServer) RefundPayment(ctx context.Context, req *paymentv1.RefundPaymentRequest) (*paymentv1.RefundPaymentResponse, error) {
	if err := ValidatePaymentID(req.GetPaymentId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request body is invalid: %v", err)
	}

	stat, err := s.service.RefundPayment(ctx, req.PaymentId, req.Reason)
	if err != nil {
		return nil, status.Errorf(GetGrpcCode(err), "failed to refund payment: %v", err)
	}

	return &paymentv1.RefundPaymentResponse{
		Status: string(stat),
	}, nil
}

func (s *PaymentServer) SuccessPayment(ctx context.Context, req *paymentv1.SuccessPaymentRequest) (*paymentv1.SuccessPaymentResponse, error) {
	if err := ValidatePaymentID(req.GetPaymentId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request body is invalid: %v", err)
	}

	stat, err := s.service.SuccessPayment(ctx, req.PaymentId)
	if err != nil {
		return nil, status.Errorf(GetGrpcCode(err), "failed to set payment status to success: %v", err)
	}

	return &paymentv1.SuccessPaymentResponse{
		Status: string(stat),
	}, nil
}

func (s *PaymentServer) HealthCheck(ctx context.Context, _ *paymentv1.HealthCheckRequest) (*paymentv1.HealthCheckResponse, error) {
	err := s.service.HealthCheck(ctx)

	return &paymentv1.HealthCheckResponse{
		DatabaseOk: !errors.Is(err, service.ErrDBUnavailable),
		BrokerOk:   !errors.Is(err, service.ErrBrokerUnavailable),
		CheckedAt:  timestamppb.Now(),
	}, nil
}

func (s *PaymentServer) ListPayments(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error) {
	if err := ValidateListPayments(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request body is invalid: %v", err)
	}

	paymentList, err := s.service.PaymentsList(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(GetGrpcCode(err), "failed to get payments list: %v", err)
	}

	return mapPaymentsToResponse(paymentList), nil
}
func (s *PaymentServer) ReversalPayment(ctx context.Context, req *paymentv1.ReversalPaymentRequest) (*paymentv1.ReversalPaymentResponse, error) {
	if err := ValidateReversalOrderReq(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request body is invalid: %v", err)
	}

	state, err := s.service.ReversalPayment(ctx, req.PaymentId, req.Amount, req.Currency)
	if err != nil {
		return nil, status.Errorf(GetGrpcCode(err), "failed to deposit payment: %v", err)
	}

	return &paymentv1.ReversalPaymentResponse{
		Status: string(state),
	}, nil
}
