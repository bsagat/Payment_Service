package domain

import (
	"time"
)

type PaymentStatus string

const (
	StatusCreated    PaymentStatus = "CREATED"    // Платёж создан, но ещё не инициирован
	StatusAuthorized PaymentStatus = "AUTHORIZED" // Авторизация завершена, средства еще не списаны
	StatusApproved   PaymentStatus = "APPROVED"   // Платёж одобрен, деньги заблокированы
	StatusDeposited  PaymentStatus = "DEPOSITED"  // Средства списаны, платёж завершён
	StatusDeclined   PaymentStatus = "DECLINED"   // Платёж отклонён банком
	StatusReversed   PaymentStatus = "REVERSED"   // Отмена клиентом до завершения платежа
	StatusRefunded   PaymentStatus = "REFUNDED"   // Возврат после завершения платежа
)

type PaymentOperation string

const (
	COFpayment = "COF_payment"
	URLpayment = "URL_payment"
)

type Payment struct {
	ID        string  // Создается на стороне брокера
	OrderID   string  // ID заказа
	UserID    string  // ID заказчика
	Broker    string  // Имя банка
	Amount    float64 // Сумма заказа
	Currency  string  //  Тип валюты (стандарт ISO)
	Operation PaymentOperation
	Status    PaymentStatus
	CreatedAt time.Time
}
