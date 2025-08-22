package ports

import (
	"context"
	"payment/internal/domain/models"
)

type Broker interface {
	CreateOrder(ctx context.Context, payment *models.Payment, returnURL string, errorURL string) (formURL string, err error)
	CreateAuthOrder(ctx context.Context, payment *models.Payment, returnURL string, errorURL string) (string, error)
	GetOrderStatus(ctx context.Context, paymentID string) (models.StatusType, error)
	GetOrderDetails(ctx context.Context, paymentID string) (models.Payment, error)
	DepositOrder(ctx context.Context, paymentID string, amount float64, currency string) error
	ReversalOrder(ctx context.Context, paymentID string, amount float64, currencyStr string) error
	RefundOrder(ctx context.Context, paymentID string, amount float64, currencyStr string) error
	Ping() error
}

type PaymentRepo interface {
	Create(ctx context.Context, transaction models.Payment) error
	Delete(ctx context.Context, paymentID string) error
	IsUnique(ctx context.Context, orderID string) (uniq bool, err error)
	GetStatus(ctx context.Context, paymentID string) (*models.PaymentStatus, error)
	MarkStatus(ctx context.Context, paymentID string, status models.StatusType) error
	Refund(ctx context.Context, paymentID, reason string, amount float64) error
	GetTransactionByPaymentID(ctx context.Context, paymentID string) (*models.Payment, error)
	UserPaymentsList(ctx context.Context, userID string, offset, limit int) ([]models.Payment, error)
	UpdateByOrderID(ctx context.Context, transaction models.Payment) error
	Ping(context.Context) error
}

type PaymentService interface {
	HealthCheck(ctx context.Context) error
	CreatePayment(ctx context.Context, orderID, userID string, amount float64, currency string, operation string, returnUrl, failUrl string) (payment models.Payment, pay_url string, err error)
	AuthPayment(ctx context.Context, orderID, userID string, amount float64, currency string, returnUrl, failUrl string) (payment models.Payment, pay_url string, err error)
	DepositPayment(context.Context, string, float64, string) (models.StatusType, error)
	GetPayment(ctx context.Context, orderID string) (models.Payment, error)
	GetPaymentStatus(ctx context.Context, orderID string) (models.StatusType, error)
	RefundPayment(ctx context.Context, orderID, reason string) (models.StatusType, error)
	SuccessPayment(ctx context.Context, orderID string) (models.StatusType, error)
	PaymentsList(ctx context.Context, userID string, pageNumber, pageSize int) ([]models.Payment, error)
	ReversalPayment(ctx context.Context, paymentID string, amount float64, currency string) (models.StatusType, error)
}
