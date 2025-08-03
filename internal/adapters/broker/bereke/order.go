package bereke

import (
	"context"
	"fmt"
	"payment/internal/domain"
	"payment/pkg/currency"
	"time"

	bma "github.com/bsagat/bereke-merchant-api"
)

// CreateOrder регистрирует новый платёж в системе Bereke.
// Возвращает ссылку на платёжную форму для редиректа пользователя.
func (c *BerekeClient) CreateOrder(ctx context.Context, payment *domain.Payment) (string, error) {
	const op = "BerekeClient.CreateOrder"

	c.log.Info(ctx, op, "order_id", payment.OrderID, "amount", payment.Amount, "currency", payment.Currency)

	orderReq := bma.RegisterOrderRequest{
		Order: bma.Order{
			OrderNumber:        payment.OrderID,
			Amount:             currency.ToMinorUnit(payment.Amount, payment.Currency),
			Currency:           currency.ToNumeric(payment.Currency),
			ReturnURL:          c.returnURL,
			FailURL:            c.errorURL,
			SessionTimeoutSecs: int(time.Minute.Seconds()) * 15,
		},
		ClientInfo: bma.ClientInfo{
			ClientId: payment.UserID,
		},
	}

	res, err := c.merchant.RegisterOrder(ctx, orderReq)
	if err != nil {
		c.log.Error(ctx, op, "err", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Проверка статуса в CreateOrder отличается
	if len(res.ErrorCode) != 0 {
		errMsg := fmt.Sprintf("response code: %s, message: %s", res.ErrorCode, res.ErrorMessage)
		c.log.Error(ctx, op, "error", errMsg)
		return "", fmt.Errorf("%s: %s", op, errMsg)
	}

	payment.ID = res.OrderID
	c.log.Info(ctx, op, "form_url", res.FormURL)

	return res.FormURL, nil
}

// GetOrderStatus получает текущий статус заказа по его ID.
func (c *BerekeClient) GetOrderStatus(ctx context.Context, paymentID string) (domain.PaymentStatus, error) {
	const op = "BerekeClient.GetOrderStatus"
	c.log.Info(ctx, op, "payment_id", paymentID)

	statusReq := bma.OrderStatusRequest{OrderID: paymentID}

	res, err := c.merchant.OrderStatus(ctx, statusReq)
	if err != nil {
		c.log.Error(ctx, op, "err", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != "0" {
		errMsg := fmt.Sprintf("response code: %s, message: %s", res.ErrorCode, res.ErrorMessage)
		c.log.Error(ctx, op, "error", errMsg)
		return "", fmt.Errorf("%s: %s", op, errMsg)
	}

	c.log.Info(ctx, op, "payment_state", res.PaymentAmountInfo.PaymentState)

	return domain.PaymentStatus(res.PaymentAmountInfo.PaymentState), nil
}

// GetOrderDetails возвращает полную информацию о заказе (в том числе дату, сумму, валюту, статус).
func (c *BerekeClient) GetOrderDetails(ctx context.Context, paymentID string) (domain.Payment, error) {
	const op = "BerekeClient.GetOrderDetails"
	c.log.Info(ctx, op, "payment_id", paymentID)

	statusReq := bma.OrderStatusRequest{OrderID: paymentID}

	res, err := c.merchant.OrderStatus(ctx, statusReq)
	if err != nil {
		c.log.Error(ctx, op, "err", err)
		return domain.Payment{}, fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != "0" {
		errMsg := fmt.Sprintf("response code: %s, message: %s", res.ErrorCode, res.ErrorMessage)
		c.log.Error(ctx, op, "error", errMsg)
		return domain.Payment{}, fmt.Errorf("%s: %s", op, errMsg)
	}

	payment := domain.Payment{
		ID:        res.OrderID,
		Broker:    Bereke_Broker,
		Amount:    currency.ConvertFromMinorUnits(res.Amount, res.Currency),
		Currency:  currency.FromString(res.Currency),
		Status:    domain.PaymentStatus(res.PaymentAmountInfo.PaymentState),
		CreatedAt: time.UnixMilli(res.Date),
		UserID:    res.BindingInfo.ClientID,
	}

	c.log.Info(ctx, op, "order", payment)

	return payment, nil
}

// ReversalOrder выполняет аннулирование (сторнирование) платежа.
// Обычно вызывается, если деньги были авторизованы, но не захвачены.
func (c *BerekeClient) ReversalOrder(ctx context.Context, orderID string, amount float64, currencyStr string) error {
	const op = "BerekeClient.ReversalOrder"
	c.log.Info(ctx, op, "order_id", orderID, "amount", amount, "currency", currencyStr)

	cancelReq := bma.ReversalOrderRequest{
		OrderID:  orderID,
		Amount:   currency.ToMinorUnit(amount, currencyStr),
		Currency: currency.ToNumeric(currencyStr),
	}

	res, err := c.merchant.ReversalOrder(ctx, cancelReq)
	if err != nil {
		c.log.Error(ctx, op, "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != "0" {
		errMsg := fmt.Sprintf("response code: %s, message: %s", res.ErrorCode, res.ErrorMessage)
		c.log.Error(ctx, op, "error", errMsg)
		return fmt.Errorf("%s: %s", op, errMsg)
	}

	c.log.Info(ctx, op, "reversal_success", true)
	return nil
}

// RefundOrder выполняет возврат средств по заказу (частично или полностью).
func (c *BerekeClient) RefundOrder(ctx context.Context, orderID string, amount float64, currencyStr string) error {
	const op = "BerekeClient.RefundOrder"
	c.log.Info(ctx, op, "order_id", orderID, "amount", amount, "currency", currencyStr)

	cancelReq := bma.RefundOrderRequest{
		OrderID:  orderID,
		Amount:   currency.ToMinorUnit(amount, currencyStr),
		Currency: currency.ToNumeric(currencyStr),
	}

	res, err := c.merchant.RefundOrder(ctx, cancelReq)
	if err != nil {
		c.log.Error(ctx, op, "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.ErrorCode != "0" {
		errMsg := fmt.Sprintf("response code: %s, message: %s", res.ErrorCode, res.ErrorMessage)
		c.log.Error(ctx, op, "error", errMsg)
		return fmt.Errorf("%s: %s", op, errMsg)
	}

	c.log.Info(ctx, op, "refund_success", true)
	return nil
}
