package repo

import "context"

func (repo *PostgresPaymentRepo) Ping(ctx context.Context) error {
	return repo.pool.Ping(ctx)
}
