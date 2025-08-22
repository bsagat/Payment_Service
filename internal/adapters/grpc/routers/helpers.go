package routers

import (
	"errors"
	"payment/internal/adapters/broker/bereke"
	paymentv1 "payment/internal/adapters/grpc/payment/v1"
	"payment/internal/adapters/repo"
	"payment/internal/domain/models"
	"payment/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GetGrpcCode(err error) codes.Code {
	switch {
	case errors.Is(err, repo.ErrPaymentNotFound), errors.Is(err, repo.ErrPaymentStatusNotFound), errors.Is(err, bereke.ErrNoSuchOrder):
		return codes.NotFound
	case errors.Is(err, repo.ErrOrderIDConflict):
		return codes.AlreadyExists
	case errors.Is(err, service.ErrUnsupportedCurrency), errors.Is(err, service.ErrPaymentNotPaid):
		return codes.InvalidArgument
	case errors.Is(err, service.ErrBrokerOperationFailed):
		return codes.FailedPrecondition
	default:
		return codes.Internal
	}
}

func mapPaymentsToResponse(payments []models.Payment) *paymentv1.ListPaymentsResponse {
	resp := &paymentv1.ListPaymentsResponse{
		Payments: make([]*paymentv1.GetPaymentResponse, 0, len(payments)),
		Total:    int32(len(payments)),
	}

	for _, p := range payments {
		resp.Payments = append(resp.Payments, &paymentv1.GetPaymentResponse{
			PaymentId: p.ID,
			OrderId:   p.OrderID,
			UserId:    p.UserID,
			Amount:    p.Amount,
			Currency:  p.Currency,
			Status:    string(p.Status),
			CreatedAt: timestamppb.New(p.CreatedAt),
			Operation: string(p.Operation),
			Broker:    p.Broker,
		})
	}

	return resp
}
