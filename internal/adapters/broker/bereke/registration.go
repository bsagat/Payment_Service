package bereke

import (
	"payment/internal/domain"
)

type Merchant struct {
	login    string
	password string
}

func NewMerchant(login, password string) *Merchant {
	return &Merchant{
		login:    login,
		password: password,
	}
}

type BerekeClient struct {
	baseURL  string
	merchant *Merchant
}

func NewClient(baseURL string, merchant *Merchant) *BerekeClient {
	return &BerekeClient{
		baseURL:  baseURL,
		merchant: merchant,
	}
}

type Broker interface {
	CreateOrder(returnURL, errorURL string) (orderID string, err error)
	GetOrderStatus(orderID string) (status string, err error)
	GetOrderDetails(orderID string) (domain.Payment, error)
	CapturePayment(orderID string, amount int64) error
	CancelOrder(orderID string) error
	RefundOrder(orderID string, amount int64) error
	Ping() error
}
