package domain

import "time"

type PaymentStatus string

const (
	StatusCreated   PaymentStatus = "CREATED"   // Заказ создан, но оплата ещё не инициирована
	StatusApproved  PaymentStatus = "APPROVED"  // Платёж одобрен: средства заблокированы, но ещё не списаны
	StatusDeposited PaymentStatus = "DEPOSITED" // Платёж завершён: средства списаны со счёта покупателя
	StatusDeclined  PaymentStatus = "DECLINED"  // Платёж отклонён банком (недостаточно средств, ошибка и т.д.)
	StatusReversed  PaymentStatus = "REVERSED"  // Платёж отменён по инициативе клиента
	StatusREFUNDED  PaymentStatus = "REFUNDED"  // Средства возвращены клиенту после успешного платежа
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
