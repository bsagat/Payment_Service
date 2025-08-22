package models

import (
	"time"
)

type PaymentOperation string

const (
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
	Status    StatusType
	CreatedAt time.Time
}
