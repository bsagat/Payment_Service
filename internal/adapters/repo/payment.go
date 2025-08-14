package repo

import (
	"context"
	"errors"
	"fmt"
	"payment/internal/domain/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPaymentRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresPaymentRepo(pool *pgxpool.Pool) *PostgresPaymentRepo {
	return &PostgresPaymentRepo{
		pool: pool,
	}
}

var (
	ErrPaymentNotFound       = errors.New("payment is not found")
	ErrOrderIDConflict       = errors.New("orderID must be unique")
	ErrPaymentStatusNotFound = errors.New("payment status info is not found")
)

// Добавляет информацию о платеже в БД
func (repo *PostgresPaymentRepo) Create(ctx context.Context, transaction models.Payment) error {
	const op = "PostgresPaymentRepo.Create"
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

func (repo *PostgresPaymentRepo) GetTransactionByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	const op = "PostgresPaymentRepo.GetTransactionByOrderID"
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

	var payment models.Payment
	if err := repo.pool.QueryRow(ctx, query, orderID).
		Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Currency, &payment.Broker, &payment.Operation, &payment.Status, &payment.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrPaymentNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &payment, nil
}

func (repo *PostgresPaymentRepo) GetTransactionByPaymentID(ctx context.Context, paymentID string) (*models.Payment, error) {
	const op = "PostgresPaymentRepo.GetTransactionByPaymentID"
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

	var payment models.Payment
	if err := repo.pool.QueryRow(ctx, query, paymentID).
		Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Currency, &payment.Broker, &payment.Operation, &payment.Status, &payment.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrPaymentNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &payment, nil
}

// Получает полную информацию о последнем статусе заказа
func (repo *PostgresPaymentRepo) GetStatus(ctx context.Context, orderID string) (*models.PaymentStatus, error) {
	const op = "PostgresPaymentRepo.GetStatus"
	query := `
	SELECT 
		s.Status, s.Order_id, s.Created_at 
	FROM 
		Status s
	WHERE 
		s.Order_id = $1
	ORDER BY 
		s.Created_at DESC
	LIMIT 1;
	`

	var status models.PaymentStatus
	if err := repo.pool.QueryRow(ctx, query, orderID).Scan(&status.Status, &status.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrPaymentStatusNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &status, nil
}

// Обновляет существующий заказ
func (repo *PostgresPaymentRepo) UpdateByOrderID(ctx context.Context, transaction models.Payment) error {
	const op = "PostgresPaymentRepo.UpdateByOrderID"
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

func (repo *PostgresPaymentRepo) Delete(ctx context.Context, orderID string) error {
	const op = "PostgresPaymentRepo.Delete"
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

func (repo *PostgresPaymentRepo) MarkStatus(ctx context.Context, orderID string, status models.StatusType) error {
	const op = "PostgresPaymentRepo.MarkStatus"

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	query := `
		UPDATE 
			Transactions
		SET 
			Current_status = $1
		WHERE
			 Order_id = $2;`

	res, err := tx.Exec(ctx, query, status, orderID)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.RowsAffected() == 0 {
		tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, ErrPaymentNotFound)
	}

	query = `
	INSERT INTO 
		Status(Order_id, Status)
	VALUES 
		($1, $2);`
	_, err = tx.Exec(ctx, query, orderID, status)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit(ctx)
}
