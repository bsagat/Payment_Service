package service

import "payment/internal/domain/models"

func IsCurrencySupported(currency string) bool {
	switch currency {
	case models.KZT, models.RUB, models.EUR, models.USD:
		return true
	default:
		return false
	}
}

func IsOperationSupported(operation string) bool {
	switch operation {
	case models.URLpayment:
		return true
	default:
		return false
	}
}
