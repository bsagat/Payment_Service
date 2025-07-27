package domain

import "time"

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
	StatusCancelled PaymentStatus = "cancelled"
)

type Payment struct {
	ID        string
	OrderID   string
	UserID    string
	Amount    float64
	Currency  string
	Status    PaymentStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
