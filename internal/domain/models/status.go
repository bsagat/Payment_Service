package models

import (
	"time"
)

type StatusType string

const (
	OrderCreated   StatusType = "CREATED"   // Заказ создан (но не оплачен)
	OrderApproved  StatusType = "APPROVED"  // Заказ одобрен (средства на счету покупателя заблокированы)
	OrderDeposited StatusType = "DEPOSITED" // Заказ завершен (деньги списаны со счета покупателя)
	OrderDeclined  StatusType = "DECLINED"  // Заказ отклонен
	OrderReversed  StatusType = "REVERSED"  // Авторизованный заказ отклонен
	OrderRefunded  StatusType = "REFUNDED"  // Возврат средств
)

type PaymentStatus struct {
	PaymentID string
	CreatedAt time.Time
	Status    string
}
