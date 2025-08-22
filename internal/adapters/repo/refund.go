package repo

import (
	"context"
	"fmt"
	"payment/internal/domain/models"
)

// Refund — выполняет возврат средств по платежу.
// В транзакции:
//  1. Сохраняет запись в Refunds,
//  2. Обновляет статус в Transactions,
//  3. Логирует новый статус в TransactionStatus.
//
// При ошибке выполняется rollback.
func (repo *PostgresPaymentRepo) Refund(ctx context.Context, paymentID, reason string, amount float64) error {
	const op = "PostgresPaymentRepo.Refund"

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// 1) Добавляем данные о возврате
	query := `
		INSERT INTO Refunds(Payment_id, Amount, Reason)
		VALUES ($1, $2, $3);`

	if _, err = tx.Exec(ctx, query, paymentID, amount, reason); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// 2) Обновляем статус заказа
	query = `
		UPDATE Transactions
		SET Current_status = $1
		WHERE Payment_id = $2;`

	res, err := tx.Exec(ctx, query, models.OrderRefunded, paymentID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}

	// 3) Добавляем новую запись о статусе заказа
	query = `
		INSERT INTO TransactionStatus(Payment_id, Status)
		VALUES ($1, $2);`

	if _, err = tx.Exec(ctx, query, paymentID, models.OrderRefunded); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit(ctx)
}
