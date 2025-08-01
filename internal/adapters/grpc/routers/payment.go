package routers

import (
	"context"
	paymentv1 "payment/internal/adapters/grpc/payment/v1"
)

type PaymentServer struct {
	paymentv1.UnimplementedPaymentServer
}

func NewPaymentServer() *PaymentServer {
	return &PaymentServer{}
}

func CreatePayment(context.Context, *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	return &paymentv1.CreatePaymentResponse{}, nil
}

func DeclinePayment(context.Context, *paymentv1.DeclinePaymentRequest) (*paymentv1.DeclinePaymentResponse, error) {
	return &paymentv1.DeclinePaymentResponse{}, nil
}

func RefundPayment(context.Context, *paymentv1.RefundPaymentRequest) (*paymentv1.RefundPaymentResponse, error) {
	return &paymentv1.RefundPaymentResponse{}, nil
}

func GetPayment(context.Context, *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	return &paymentv1.GetPaymentResponse{}, nil
}

func COFPayment(context.Context, *paymentv1.COFPaymentRequest) (*paymentv1.COFPaymentResponse, error) {
	return &paymentv1.COFPaymentResponse{}, nil
}

func SuccessPayment(context.Context, *paymentv1.SuccessPaymentRequest) (*paymentv1.SuccessPaymentResponse, error) {
	return &paymentv1.SuccessPaymentResponse{}, nil
}
