package models

import "time"

type StatusType string

const (
	StatusCreated    StatusType = "CREATED"    // Платёж создан, но ещё не инициирован
	StatusAuthorized StatusType = "AUTHORIZED" // Авторизация завершена, средства еще не списаны
	StatusApproved   StatusType = "APPROVED"   // Платёж одобрен, деньги заблокированы
	StatusDeposited  StatusType = "DEPOSITED"  // Средства списаны, платёж завершён
	StatusDeclined   StatusType = "DECLINED"   // Платёж отклонён банком
	StatusRefunded   StatusType = "REFUNDED"   // Возврат после завершения платежа
)

type PaymentStatus struct {
	OrderID   string
	CreatedAt time.Time
	Status    string
}
