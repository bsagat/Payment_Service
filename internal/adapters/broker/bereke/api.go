package bereke

import (
	"context"
	"fmt"
	"payment/internal/domain"
	"payment/pkg/logger"

	bma "github.com/bsagat/bereke-merchant-api"
)

var Bereke_Broker string = "BEREKE"

type BerekeClient struct {
	merchant bma.API
	log      logger.Logger

	returnURL string
	errorURL  string
}

type Broker interface {
	CreateOrder(ctx context.Context, payment *domain.Payment) (formURL string, err error)
	GetOrderStatus(ctx context.Context, paymentID string) (domain.PaymentStatus, error)
	GetOrderDetails(ctx context.Context, paymentID string) (domain.Payment, error)
	ReversalOrder(ctx context.Context, orderID string, amount float64, currencyStr string) error
	RefundOrder(ctx context.Context, orderID string, amount float64, currencyStr string) error
	Ping() error
}

func NewClient(returnURL, errorURL string, api_key string, mode bma.Mode) (Broker, error) {
	api, err := bma.NewWithToken(api_key, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create new bma client: %v", err)
	}

	return &BerekeClient{
		returnURL: returnURL,
		errorURL:  errorURL,
		merchant:  api,
	}, nil
}

func (c *BerekeClient) Ping() error {
	return c.merchant.Ping()
}
