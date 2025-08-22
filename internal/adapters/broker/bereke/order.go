package bereke

import (
	"context"
	"errors"
	"fmt"
	"payment/internal/domain/models"
	"time"

	money "github.com/bsagat/bereke-merchant-api/currency"
	"github.com/bsagat/bereke-merchant-api/models/code"
)

var (
	ErrNoSuchOrder         = errors.New("no such order")
	ErrOperationImpossible = errors.New("impossible for current transaction state")
)

func (c *BerekeClient) CreateOrder(ctx context.Context, payment *models.Payment, returnURL, errorURL string) (string, error) {
	const op = "BerekeClient.CreateOrder"

	res, err := c.merchant.RegisterOrderByNumber(ctx, payment.OrderID, payment.Amount, money.ToNumeric(payment.Currency), returnURL, errorURL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != code.Success {
		return "", fmt.Errorf("%s: %d %s", op, res.ErrorCode, res.ErrorMessage)
	}

	payment.ID = res.OrderID
	payment.Broker = Bereke_Broker

	return res.FormURL, nil
}

func (c *BerekeClient) CreateAuthOrder(ctx context.Context, payment *models.Payment, returnURL, errorURL string) (string, error) {
	const op = "BerekeClient.CreateAuthOrder"

	res, err := c.merchant.AuthOrderByNumber(ctx, payment.OrderID, payment.Amount, money.ToNumeric(payment.Currency), returnURL, errorURL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != code.Success {
		return "", fmt.Errorf("%s: %d %s", op, res.ErrorCode, res.ErrorMessage)
	}

	payment.ID = res.OrderID
	payment.Broker = Bereke_Broker

	return res.FormURL, nil
}

func (c *BerekeClient) GetOrderStatus(ctx context.Context, paymentID string) (models.StatusType, error) {
	const op = "BerekeClient.GetOrderStatus"

	res, err := c.merchant.GetOrderStatusByID(ctx, paymentID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != code.Success {
		return "", fmt.Errorf("%s: %d %s", op, res.ErrorCode, res.ErrorMessage)
	}

	return models.StatusType(res.PaymentAmountInfo.PaymentState), nil
}

func (c *BerekeClient) GetOrderDetails(ctx context.Context, paymentID string) (models.Payment, error) {
	const op = "BerekeClient.GetOrderDetails"

	res, err := c.merchant.GetOrderStatusByID(ctx, paymentID)
	if err != nil {
		return models.Payment{}, fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != code.Success {
		return models.Payment{}, fmt.Errorf("%s: %d %s", op, res.ErrorCode, res.ErrorMessage)
	}

	payment := models.Payment{
		ID:        res.OrderID,
		Broker:    Bereke_Broker,
		Amount:    res.Amount,
		Currency:  money.ToAlpha(res.Currency),
		Status:    models.StatusType(res.PaymentAmountInfo.PaymentState),
		CreatedAt: time.UnixMilli(res.Date),
		UserID:    res.BindingInfo.ClientID,
	}

	return payment, nil
}

func (c *BerekeClient) ReversalOrder(ctx context.Context, orderID string, amount float64, currency string) error {
	const op = "BerekeClient.ReversalOrder"

	res, err := c.merchant.ReversalOrderByID(ctx, amount, money.ToNumeric(currency), orderID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != code.Success {
		return fmt.Errorf("%s: %d %s", op, res.ErrorCode, res.ErrorMessage)
	}

	return nil
}

func (c *BerekeClient) RefundOrder(ctx context.Context, orderID string, amount float64, currency string) error {
	const op = "BerekeClient.RefundOrder"

	res, err := c.merchant.RefundOrderByID(ctx, amount, money.ToNumeric(currency), orderID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != code.Success {
		return fmt.Errorf("%s: %d %s", op, res.ErrorCode, res.ErrorMessage)
	}

	return nil
}

func (c *BerekeClient) DepositOrder(ctx context.Context, orderID string, amount float64, currency string) error {
	const op = "BerekeClient.DepositOrder"

	res, err := c.merchant.DepositOrderByNumber(ctx, orderID, amount, money.ToNumeric(currency))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != code.Success {
		return fmt.Errorf("%s: %d %s", op, res.ErrorCode, res.ErrorMessage)
	}

	return nil
}
