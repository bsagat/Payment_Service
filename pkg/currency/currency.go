package currency

import (
	"math"
	"strconv"
)

const (
	KZT string = "KZT" // 398
	USD string = "USD" // 840
	RUB string = "RUB" // 643
	EUR string = "EUR" // 978
)

// ISO currency map
var (
	alphaToNumeric = map[string]int{
		KZT: 398,
		USD: 840,
		RUB: 643,
		EUR: 978,
	}

	numericToAlpha = map[int]string{
		398: KZT,
		840: USD,
		643: RUB,
		978: EUR,
	}
)

// Конвертирует ISO currency code в string
func ToAlpha(code int) string {
	return numericToAlpha[code]
}

// Конвертирует ISO currency code в int
func ToNumeric(code string) int {
	return alphaToNumeric[code]
}

// Выполняет обработку string ISO code ("398", "KZT") и возвращает string
func FromString(value string) string {
	// Try as numeric
	if num, err := strconv.Atoi(value); err == nil {
		return ToAlpha(num)
	}
	// Try as alpha
	return value
}

func ToMinorUnit(amount float64, currency string) int {
	switch currency {
	case KZT:
		return int(math.Round(amount)) // без копеек
	case RUB, USD, EUR:
		return int(math.Round(amount * 100)) // 2 знака
	default:
		// по умолчанию считаем 2 знака
		return int(math.Round(amount * 100))
	}
}

func ConvertFromMinorUnits(amount int, currency string) float64 {
	switch currency {
	case KZT:
		return float64(amount) // без копеек
	case RUB, USD, EUR:
		return float64(amount / 100) // 2 знака
	default:
		// по умолчанию считаем 2 знака
		return float64(amount / 100) // 2 знака
	}
}
