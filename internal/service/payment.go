package service

import (
	"context"
	"errors"
	"payment/internal/domain/action"
	"payment/internal/domain/models"
	"payment/internal/domain/ports"
	"payment/pkg/logger"
)

type PaymentService struct {
	broker ports.Broker
	repo   ports.PaymentRepo

	log logger.Logger
}

func NewPaymentService(broker ports.Broker, repo ports.PaymentRepo, log logger.Logger) *PaymentService {
	return &PaymentService{
		broker: broker,
		repo:   repo,
		log:    log,
	}
}

var (
	ErrDBUnavailable     = errors.New("database unavailable")
	ErrBrokerUnavailable = errors.New("broker unavailable")
)

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

// Создает платеж на основе введенных аргументов, возвращая ссылку для оплаты
func (s *PaymentService) CreatePayment(
	ctx context.Context,
	orderID, userID string,
	amount float64,
	currency string,
	returnUrl, failUrl string,
) (models.Payment, string, error) {
	s.log.Debug(ctx,
		"create_payment_begin",
		"Begin to create new payment",
		"order_id", orderID,
		"user_id", userID,
		"amount", amount,
		"currency", currency,
		"return_url", returnUrl,
		"fail_url", failUrl,
	)

	// Создаем платеж
	payment := models.Payment{
		OrderID:  orderID,
		UserID:   userID,
		Amount:   amount,
		Currency: currency,
	}

	// Создаем платеж на стороне брокера (шлюз)
	formUrl, err := s.broker.CreateOrder(ctx, &payment, returnUrl, failUrl)
	if err != nil {
		s.log.Error(ctx, action.PaymentTransactionFail, err, "Failed to create new payment")
		return models.Payment{}, "", err
	}

	// Добавляем платеж в бд
	if err := s.repo.Create(ctx, payment); err != nil {
		s.log.Error(ctx, action.DbTransactionFailed, err, "Failed to create payment data")

		// Отменяем платеж со стороны шлюза
		if err := s.broker.CancelOrder(ctx, orderID); err != nil {
			s.log.Error(ctx, action.PaymentTransactionFail, err, "Failed to cancel order")
		}
		return models.Payment{}, "", err
	}

	s.log.Info(ctx, "create_payment_success", "Succesfully created new payment", "payment_ID", payment.ID, "broker", payment.Broker)
	return payment, formUrl, nil
}

func (s *PaymentService) GetPayment(
	ctx context.Context,
	orderID string,
) (models.Payment, error) {
	s.log.Debug(ctx,
		"get_payment_begin",
		"Payment data retrieve started",
		"order_id", orderID,
	)

	// Получаем данные о платеже через бд
	payment, err := s.repo.GetTransactionByOrderID(ctx, orderID)
	if err != nil {
		s.log.Error(ctx, action.DbTransactionFailed, err, "Failed to get payment data")
		return models.Payment{}, err
	}

	s.log.Debug(ctx,
		"get_payment_success",
		"Payment data retrieved succesfully",
		"order_id", orderID,
		"payment_id", payment.ID,
		"amount", payment.Amount,
		"currency", payment.Currency,
	)
	return *payment, nil
}

func (s *PaymentService) GetPaymentStatus(ctx context.Context, orderID string) (models.StatusType, error) {
	s.log.Debug(ctx,
		"get_payment_status_begin",
		"Payment status info retrieve started",
		"order_id", orderID,
	)

	// Получаем информацию о статусе через бд
	paymentStatus, err := s.repo.GetStatus(ctx, orderID)
	if err != nil {
		s.log.Error(ctx, action.DbTransactionFailed, err, "Failed to get payment status info")
		return "", err
	}

	s.log.Info(ctx,
		"get_payment_status_success",
		"Payment status info retrieve finished",
		"status", paymentStatus.Status,
		"order_id", orderID,
	)
	return models.StatusType(paymentStatus.Status), nil
}

func (s *PaymentService) RefundPayment(ctx context.Context, orderID, reason string) error {
	s.log.Debug(ctx,
		"refund_payment_begin",
		"Payment refund has been started",
		"order_id", orderID,
	)

	// Получаем данные о платеже
	payment, err := s.repo.GetTransactionByOrderID(ctx, orderID)
	if err != nil {
		s.log.Error(ctx, action.DbTransactionFailed, err, "Failed to get transaction", "orderID", orderID)
		return err
	}

	// Изменяем статус, сохраняя старый для дальнейшего использования
	oldStatus := payment.Status
	if err := s.repo.MarkStatus(ctx, orderID, models.StatusRefunded); err != nil {
		s.log.Error(ctx, action.DbTransactionFailed, err, "Failed to mark transaction status", "orderID", orderID, "status", models.StatusRefunded)
		return err
	}

	// Возвращаем средства пользователю
	if err := s.broker.RefundOrder(ctx, orderID, payment.Amount, payment.Currency); err != nil {
		s.log.Error(ctx, action.DbTransactionFailed, err, "Failed to get transaction", "orderID", orderID)

		// При неудачной попытке, возвращаем старый статус
		if err := s.repo.MarkStatus(ctx, orderID, oldStatus); err != nil {
			s.log.Error(ctx, action.DbTransactionFailed, err, "Failed to mark transaction status", "orderID", orderID, "status", oldStatus)
			return err
		}

		return err
	}

	s.log.Info(ctx,
		"refund_payment_success",
		"Payment refund has been succesfully finished",
		"order_id", orderID,
	)
	return nil
}

func (s *PaymentService) SuccessPayment(ctx context.Context, orderID string) error {
	s.log.Debug(ctx,
		"success_payment_begin",
		"Changing payment status to success",
		"order_id", orderID,
	)

	if err := s.repo.MarkStatus(ctx, orderID, models.StatusDeposited); err != nil {
		s.log.Error(ctx, action.DbTransactionFailed, err, "Failed to mark status as success", "order_ID", orderID)
		return err
	}

	s.log.Info(ctx,
		"success_payment_finished",
		"Payment status has been changed",
		"status", models.StatusDeposited,
		"order_id", orderID,
	)
	return nil
}
