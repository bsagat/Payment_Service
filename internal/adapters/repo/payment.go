package repo

import (
	"context"
	"errors"
	"fmt"
	"payment/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{
		pool: pool,
	}
}

var (
	ErrPaymentNotFound       = errors.New("payment is not found")
	ErrPaymentStatusNotFound = errors.New("payment status info is not found")
)

// Добавляет информацию о платеже в БД
func (repo *PaymentRepo) Create(ctx context.Context, transaction domain.Payment) error {
	const op = "PaymentRepo.Create"
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	query := `
	INSERT INTO 
		Transactions(Payment_id, User_id, Order_id, Amount, Currency, Broker, Operation)
	VALUES 
		($1, $2, $3, $4, $5, $6, $7);
	`
	_, err = tx.Exec(ctx, query, transaction.ID, transaction.UserID, transaction.OrderID, transaction.Amount, transaction.Currency, transaction.Broker, transaction.Operation)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	query = `
	INSERT INTO 
		Status(Order_id, Status)
	VALUES 
		($1, $2);`
	_, err = tx.Exec(ctx, query, transaction.OrderID, transaction.Status)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit(ctx)
}

func (repo *PaymentRepo) GetTransactionByOrderID(ctx context.Context, orderID string) (domain.Payment, error) {
	const op = "PaymentRepo.GetTransactionByOrderID"
	query := `
		SELECT 
			f.Payment_id,
			f.User_id,
			f.Order_id,
			f.Amount,
			f.Currency,
			f.Broker,
			f.Operation,
			s.Status,
			s.Created_at
		FROM 
			Transactions f 
		INNER JOIN Status s ON s.Order_id = f.Order_id
		WHERE 
			s.Order_id = $1;
		ORDER BY 
			s.Created_at DESC
		LIMIT 1;`

	var payment domain.Payment
	if err := repo.pool.QueryRow(ctx, query, orderID).
		Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Currency, &payment.Broker, &payment.Operation, &payment.Status, &payment.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Payment{}, fmt.Errorf("%s: %w", op, ErrPaymentNotFound)
		}
		return domain.Payment{}, fmt.Errorf("%s: %w", op, err)
	}
	return payment, nil
}

func (repo *PaymentRepo) GetTransactionByPaymentID(ctx context.Context, paymentID string) (domain.Payment, error) {
	const op = "PaymentRepo.GetTransactionByPaymentID"
	query := `
		SELECT 
			f.Payment_id,
			f.User_id,
			f.Order_id,
			f.Amount,
			f.Currency,
			f.Broker,
			f.Operation,
			s.Status,
			s.Created_at
		FROM 
			Transactions f 
		INNER JOIN Status s ON s.Order_id = f.Order_id
		WHERE 
			s.Payment_id = $1;
		ORDER BY 
			s.Created_at DESC
		LIMIT 1;`

	var payment domain.Payment
	if err := repo.pool.QueryRow(ctx, query, paymentID).
		Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Currency, &payment.Broker, &payment.Operation, &payment.Status, &payment.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Payment{}, fmt.Errorf("%s: %w", op, ErrPaymentNotFound)
		}
		return domain.Payment{}, fmt.Errorf("%s: %w", op, err)
	}
	return payment, nil
}

// Получает полную информацию о последнем статусе заказа
func (repo *PaymentRepo) GetStatus(ctx context.Context, orderID string) (domain.PaymentStatus, error) {
	const op = "PaymentRepo.GetStatus"
	query := `
	SELECT 
		s.Status 
	FROM 
		Status s
	WHERE 
		s.Order_id = $1
	ORDER BY 
		s.Created_at DESC
	LIMIT 1;
	`

	var status domain.PaymentStatus
	if err := repo.pool.QueryRow(ctx, query, orderID).Scan(&status); err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("%s: %w", op, ErrPaymentStatusNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return status, nil
}

// Обновляет существующий заказ
func (repo *PaymentRepo) UpdateByOrderID(ctx context.Context, transaction domain.Payment) error {
	const op = "PaymentRepo.UpdateByOrderID"
	query := `
		UPDATE 
			Transactions
		SET 
			Payment_id = $1,
			User_id = $2,
			Amount = $3,
			Currency = $4,
			Broker = $5,
			Operation = $6
		WHERE 
			Order_id = $7;`

	res, err := repo.pool.Exec(ctx, query, transaction.ID, transaction.UserID, transaction.Amount, transaction.Currency, transaction.Broker, transaction.Operation, transaction.OrderID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, ErrPaymentNotFound)
	}
	return nil
}

func (repo *PaymentRepo) Delete(ctx context.Context, orderID string) error {
	const op = "PaymentRepo.Delete"
	query := `
		DELETE FROM 
			Transactions
		WHERE 
			Order_id = $1;`
	res, err := repo.pool.Exec(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, ErrPaymentNotFound)
	}
	return nil
}
