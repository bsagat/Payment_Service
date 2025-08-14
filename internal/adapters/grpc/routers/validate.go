package routers

import (
	"errors"
	"fmt"
	paymentv1 "payment/internal/adapters/grpc/payment/v1"
)

func ValidateCreateOrderReq(req *paymentv1.CreatePaymentRequest) error {
	if req.GetAmount() <= 0 {
		return fmt.Errorf("amount %.2f must be greater than 0", req.Amount)
	}

	if req.GetCurrency() == "" {
		return errors.New("currency field is empty")
	}

	if req.GetErrorUrl() == "" {
		return errors.New("errorUrl field is empty")
	}

	if req.GetOrderId() == "" {
		return errors.New("orderID field is empty")
	}

	if req.GetReturnUrl() == "" {
		return errors.New("returnUrl field is empty")
	}

	if req.GetUserId() == "" {
		return errors.New("userID field is empty")
	}
	return nil
}

func ValidatePaymentID(paymentID string) error {
	if paymentID == "" {
		return errors.New("paymentID field is empty")
	}

	return nil
}
