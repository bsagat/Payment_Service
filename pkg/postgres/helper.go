package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const UniqueViolationCode = "23505"

// IsUniqueViolation проверяет, является ли ошибка нарушением уникального ограничения.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == UniqueViolationCode
}
