package ports

import (
	"context"
	"payment/internal/domain/models"
)

type Broker interface {
	CreateOrder(ctx context.Context, payment *models.Payment, returnURL string, errorURL string) (formURL string, err error)
	GetOrderStatus(ctx context.Context, paymentID string) (models.StatusType, error)
	GetOrderDetails(ctx context.Context, paymentID string) (models.Payment, error)
	ReversalOrder(ctx context.Context, orderID string, amount float64, currencyStr string) error
	RefundOrder(ctx context.Context, orderID string, amount float64, currencyStr string) error
	CancelOrder(ctx context.Context, orderId string) error
	Ping() error
}

type PaymentRepo interface {
	Create(ctx context.Context, transaction models.Payment) error
	Delete(ctx context.Context, orderID string) error
	GetStatus(ctx context.Context, orderID string) (*models.PaymentStatus, error)
	MarkStatus(ctx context.Context, orderID string, status models.StatusType) error
	GetTransactionByOrderID(ctx context.Context, orderID string) (*models.Payment, error)
	GetTransactionByPaymentID(ctx context.Context, paymentID string) (*models.Payment, error)
	UpdateByOrderID(ctx context.Context, transaction models.Payment) error
	Ping(context.Context) error
}

type PaymentService interface {
	HealthCheck(ctx context.Context) error
	CreatePayment(ctx context.Context, orderID, userID string, amount float64, currency string, returnUrl, failUrl string) (models.Payment, string, error)
	GetPayment(ctx context.Context, orderID string) (models.Payment, error)
	GetPaymentStatus(ctx context.Context, orderID string) (models.StatusType, error)
	RefundPayment(ctx context.Context, orderID, reason string) error
	SuccessPayment(ctx context.Context, orderID string) error
}
