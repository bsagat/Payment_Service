package service

import (
	"context"
	"errors"
	"fmt"
	"payment/internal/adapters/repo"
	"payment/internal/domain/action"
	"payment/internal/domain/models"
	"payment/internal/domain/ports"
	"payment/pkg/logger"
)

type PaymentService struct {
	broker ports.Broker
	repo   ports.PaymentRepo
	log    logger.Logger
}

func NewPaymentService(broker ports.Broker, repo ports.PaymentRepo, log logger.Logger) *PaymentService {
	return &PaymentService{
		broker: broker,
		repo:   repo,
		log:    log,
	}
}

var (
	ErrDBUnavailable         = errors.New("database unavailable")
	ErrOrderNotUnique        = errors.New("order ID is not unique")
	ErrBrokerUnavailable     = errors.New("broker unavailable")
	ErrBrokerOperationFailed = errors.New("broker operation failed")
	ErrUnsupportedCurrency   = errors.New("unsupported currency")
	ErrUnsupportedOperation  = errors.New("unsupported operation")
	ErrPaymentNotPaid        = errors.New("payment is not paid yet")
)

// HealthCheck — проверка доступности БД и брокера.
func (s *PaymentService) HealthCheck(ctx context.Context) error {
	var errsList []error

	if err := s.repo.Ping(ctx); err != nil {
		s.log.Error(ctx, action.HealthCheck, err, "database is unavailable")
		errsList = append(errsList, ErrDBUnavailable)
	}

	if err := s.broker.Ping(); err != nil {
		s.log.Error(ctx, action.HealthCheck, err, "broker is unavailable")
		errsList = append(errsList, ErrBrokerUnavailable)
	}

	if len(errsList) == 0 {
		return nil
	}
	return errors.Join(errsList...)
}

// CreatePayment — создаёт платёж и возвращает URL оплаты.
func (s *PaymentService) CreatePayment(
	ctx context.Context,
	orderID, userID string,
	amount float64,
	currency string,
	operation string,
	returnURL, failURL string,
) (models.Payment, string, error) {
	l := s.log.With(
		"order_id", orderID,
		"user_id", userID,
		"amount", amount,
		"currency", currency,
		"operation", operation,
	)
	l.Debug(ctx, action.CreatePayment, "begin")

	// Валидация
	if !IsCurrencySupported(currency) {
		l.Error(ctx, action.ValidationFailed, ErrUnsupportedCurrency, "currency is not supported")
		return models.Payment{}, "", ErrUnsupportedCurrency
	}
	if !IsOperationSupported(operation) {
		l.Error(ctx, action.ValidationFailed, ErrUnsupportedOperation, "operation is not supported")
		return models.Payment{}, "", ErrUnsupportedOperation
	}

	uniq, err := s.repo.IsUnique(ctx, orderID)
	if err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "Failed to check order uniqueness")
		return models.Payment{}, "", err
	}

	if !uniq {
		l.Error(ctx, action.ValidationFailed, repo.ErrOrderIDConflict, "order ID is not unique")
		return models.Payment{}, "", repo.ErrOrderIDConflict
	}

	// Локальная модель
	payment := models.Payment{
		OrderID:   orderID,
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Operation: models.PaymentOperation(operation),
		Status:    models.OrderCreated,
	}

	// Создание заказа у брокера
	formURL, err := s.broker.CreateOrder(ctx, &payment, returnURL, failURL)
	if err != nil {
		l.Error(ctx, action.PaymentTransactionFail, err, "failed to create payment at broker")
		return models.Payment{}, "", fmt.Errorf("%w: %v", ErrBrokerOperationFailed, err)
	}

	// Сохранение в БД
	if err := s.repo.Create(ctx, payment); err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "failed to persist payment")
		return models.Payment{}, "", err
	}

	s.log.With("payment_id", payment.ID, "broker", payment.Broker).
		Info(ctx, action.CreatePayment, "success")
	return payment, formURL, nil
}

// GetPayment — возвращает платёж по orderID.
func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (models.Payment, error) {
	l := s.log.With("order_id", paymentID)
	l.Debug(ctx, action.GetPayment, "begin")

	payment, err := s.repo.GetTransactionByPaymentID(ctx, paymentID)
	if err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "failed to get payment")
		return models.Payment{}, err
	}

	s.log.With("payment_id", payment.ID, "amount", payment.Amount, "currency", payment.Currency).
		Debug(ctx, action.GetPayment, "success")
	return *payment, nil
}

// GetPaymentStatus — возвращает статус платежа.
func (s *PaymentService) GetPaymentStatus(ctx context.Context, paymentID string) (models.StatusType, error) {
	l := s.log.With("order_id", paymentID)
	l.Debug(ctx, action.GetPaymentStatus, "begin")

	status, err := s.repo.GetStatus(ctx, paymentID)
	if err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "failed to get payment status")
		return "", err
	}

	s.log.With("status", status.Status).Info(ctx, action.GetPaymentStatus, "success")
	return models.StatusType(status.Status), nil
}

// RefundPayment — инициирует возврат и меняет статус.
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID, reason string) (models.StatusType, error) {
	l := s.log.With("payment_id", paymentID, "reason", reason)
	l.Debug(ctx, action.RefundPayment, "begin")

	payment, err := s.repo.GetTransactionByPaymentID(ctx, paymentID)
	if err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "failed to load payment")
		return "", err
	}

	if err := s.broker.RefundOrder(ctx, paymentID, payment.Amount, payment.Currency); err != nil {
		l.Error(ctx, action.PaymentTransactionFail, err, "broker refund failed")
		return "", fmt.Errorf("%w: %v", ErrBrokerOperationFailed, err)
	}

	if err := s.repo.Refund(ctx, paymentID, reason, payment.Amount); err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "failed to mark refund in db")
		return "", err
	}

	s.log.With("refunded_amount", payment.Amount).Info(ctx, action.RefundPayment, "success")
	return models.OrderRefunded, nil
}

// SuccessPayment — помечает платёж как успешный (DEPOSITED).
func (s *PaymentService) SuccessPayment(ctx context.Context, paymentID string) (models.StatusType, error) {
	l := s.log.With("payment_id", paymentID)
	l.Debug(ctx, action.SuccessPayment, "begin")

	status, err := s.broker.GetOrderStatus(ctx, paymentID)
	if err != nil {
		l.Error(ctx, action.PaymentTransactionFail, err, "failed to get payment status")
		return "", fmt.Errorf("%w: %v", ErrBrokerOperationFailed, err)
	}

	if status != models.OrderApproved && status != models.OrderDeposited {
		l.Error(ctx, action.ValidationFailed, ErrPaymentNotPaid, "Order is not paid yet")
		return "", ErrPaymentNotPaid
	}

	if err := s.repo.MarkStatus(ctx, paymentID, status); err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "failed to mark deposited")
		return "", err
	}

	l.Info(ctx, action.SuccessPayment, "success")
	return status, nil
}

// PaymentsList — список платежей пользователя с пагинацией.
func (s *PaymentService) PaymentsList(ctx context.Context, userID string, pageNum, pageSize int) ([]models.Payment, error) {
	offset := (pageNum - 1) * pageSize
	l := s.log.With("user_id", userID, "page", pageNum, "page_size", pageSize, "offset", offset)
	l.Debug(ctx, action.ListPayments, "begin")

	list, err := s.repo.UserPaymentsList(ctx, userID, offset, pageSize)
	if err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "failed to get user payments list")
		return nil, err
	}

	l.Info(ctx, action.ListPayments, "success")
	return list, nil
}

// AuthPayment — создаёт авторизованный платёж (hold).
func (s *PaymentService) AuthPayment(
	ctx context.Context,
	orderID, userID string,
	amount float64,
	currency string,
	returnURL, failURL string,
) (models.Payment, string, error) {
	l := s.log.With(
		"order_id", orderID,
		"user_id", userID,
		"amount", amount,
		"currency", currency,
	)
	l.Debug(ctx, action.AuthPayment, "begin")

	if !IsCurrencySupported(currency) {
		l.Error(ctx, action.ValidationFailed, ErrUnsupportedCurrency, "currency is not supported")
		return models.Payment{}, "", ErrUnsupportedCurrency
	}

	uniq, err := s.repo.IsUnique(ctx, orderID)
	if err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "Failed to check order uniqueness")
		return models.Payment{}, "", err
	}

	if !uniq {
		l.Error(ctx, action.ValidationFailed, repo.ErrOrderIDConflict, "order ID is not unique")
		return models.Payment{}, "", repo.ErrOrderIDConflict
	}

	payment := models.Payment{
		OrderID:   orderID,
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Operation: models.URLpayment,
		Status:    models.OrderCreated,
	}

	formURL, err := s.broker.CreateAuthOrder(ctx, &payment, returnURL, failURL)
	if err != nil {
		l.Error(ctx, action.PaymentTransactionFail, err, "failed to create auth order at broker")
		return models.Payment{}, "", fmt.Errorf("%w: %v", ErrBrokerOperationFailed, err)
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "failed to persist auth payment")
		return models.Payment{}, "", err
	}

	s.log.With("payment_id", payment.ID, "broker", payment.Broker).
		Info(ctx, action.AuthPayment, "success")
	return payment, formURL, nil
}

// DepositPayment — списывает (capture) ранее авторизованные средства.
func (s *PaymentService) DepositPayment(ctx context.Context, paymentID string, amount float64, currency string) (models.StatusType, error) {
	l := s.log.With("payment_id", paymentID, "amount", amount, "currency", currency)
	l.Debug(ctx, action.DepositPayment, "begin")

	// Списывание полной суммы
	if amount == 0 {
		payment, err := s.repo.GetTransactionByPaymentID(ctx, paymentID)
		if err != nil {
			l.Error(ctx, action.DbTransactionFailed, err, "failed to load payment")
			return "", err
		}

		amount = payment.Amount
		currency = payment.Currency
	}

	// Инициируем списание у брокера
	if err := s.broker.DepositOrder(ctx, paymentID, amount, currency); err != nil {
		l.Error(ctx, action.PaymentTransactionFail, err, "broker deposit failed")
		return "", fmt.Errorf("%w: %v", ErrBrokerOperationFailed, err)
	}

	// Обновляем статус
	if err := s.repo.MarkStatus(ctx, paymentID, models.OrderDeposited); err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "failed to mark deposited in db")
		return "", err
	}

	l.Info(ctx, action.DepositPayment, "success")
	return models.OrderDeposited, nil
}

func (s *PaymentService) ReversalPayment(ctx context.Context, paymentID string, amount float64, currency string) (models.StatusType, error) {
	l := s.log.With("payment_id", paymentID, "amount", amount, "currency", currency)
	l.Debug(ctx, action.ReversePayment, "begin")

	// Списывание полной суммы
	if amount == 0 {
		payment, err := s.repo.GetTransactionByPaymentID(ctx, paymentID)
		if err != nil {
			l.Error(ctx, action.DbTransactionFailed, err, "failed to load payment")
			return "", err
		}

		amount = payment.Amount
		currency = payment.Currency
	}

	// Инициируем реверсирование средств
	if err := s.broker.ReversalOrder(ctx, paymentID, amount, currency); err != nil {
		l.Error(ctx, action.PaymentTransactionFail, err, "broker reversal failed")
		return "", fmt.Errorf("%w: %v", ErrBrokerOperationFailed, err)
	}

	// Обновляем статус
	if err := s.repo.MarkStatus(ctx, paymentID, models.OrderReversed); err != nil {
		l.Error(ctx, action.DbTransactionFailed, err, "failed to mark reversed in db")
		return "", err
	}

	l.Info(ctx, action.ReversePayment, "success")
	return models.OrderReversed, nil
}
