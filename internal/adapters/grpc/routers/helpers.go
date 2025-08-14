package routers

import (
	"errors"
	"payment/internal/adapters/repo"

	"google.golang.org/grpc/codes"
)

func GetGrpcCode(err error) codes.Code {
	switch {
	case errors.Is(err, repo.ErrPaymentNotFound), errors.Is(err, repo.ErrPaymentStatusNotFound):
		return codes.NotFound
	case errors.Is(err, repo.ErrOrderIDConflict):
		return codes.AlreadyExists
	default:
		return codes.Internal
	}
}
