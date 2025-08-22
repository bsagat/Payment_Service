package repo

import (
	"context"
	"errors"
	"fmt"
	"payment/internal/domain/models"
	"payment/pkg/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPaymentRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresPaymentRepo(pool *pgxpool.Pool) *PostgresPaymentRepo {
	return &PostgresPaymentRepo{pool: pool}
}

var (
	ErrPaymentNotFound       = errors.New("payment is not found")
	ErrOrderIDConflict       = errors.New("orderID must be unique")
	ErrPaymentStatusNotFound = errors.New("payment status info is not found")
)

// Добавляет информацию о платеже в БД
func (repo *PostgresPaymentRepo) Create(ctx context.Context, transaction models.Payment) (err error) {
	const op = "PostgresPaymentRepo.Create"

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO Transactions(Payment_id, User_id, Order_id, Amount, Currency, Broker, Operation)
		VALUES ($1, $2, $3, $4, $5, $6, $7);`
	_, err = tx.Exec(ctx, query,
		transaction.ID, transaction.UserID, transaction.OrderID,
		transaction.Amount, transaction.Currency, transaction.Broker, transaction.Operation)
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return ErrOrderIDConflict
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// сохраняем статус по Payment_id
	query = `
		INSERT INTO TransactionStatus(Payment_id, Status)
		VALUES ($1, $2);`
	_, err = tx.Exec(ctx, query, transaction.ID, transaction.Status)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit(ctx)
}

// Проверяет уникальность OrderID
func (repo *PostgresPaymentRepo) IsUnique(ctx context.Context, orderID string) (bool, error) {
	const op = "PostgresPaymentRepo.IsUnique"
	query := `
		SELECT COUNT(*) = 0
		FROM Transactions
		WHERE Order_id = $1;`

	var unique bool
	if err := repo.pool.QueryRow(ctx, query, orderID).Scan(&unique); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return unique, nil
}

// Получает транзакцию по OrderID
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
		INNER JOIN 
			TransactionStatus s ON s.Payment_id = f.Payment_id
		WHERE 	
			f.Order_id = $1
		ORDER BY 
			s.Created_at DESC
		LIMIT 1;`

	var payment models.Payment
	if err := repo.pool.QueryRow(ctx, query, orderID).
		Scan(&payment.ID, &payment.UserID, &payment.OrderID,
			&payment.Amount, &payment.Currency, &payment.Broker,
			&payment.Operation, &payment.Status, &payment.CreatedAt); err != nil {

		if err == pgx.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &payment, nil
}

// Получает транзакцию по PaymentID
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
		INNER JOIN TransactionStatus s ON s.Payment_id = f.Payment_id
		WHERE 
			f.Payment_id = $1
		ORDER BY 
			s.Created_at DESC
		LIMIT 1;`

	var payment models.Payment
	if err := repo.pool.QueryRow(ctx, query, paymentID).
		Scan(&payment.ID, &payment.UserID, &payment.OrderID,
			&payment.Amount, &payment.Currency, &payment.Broker,
			&payment.Operation, &payment.Status, &payment.CreatedAt); err != nil {

		if err == pgx.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &payment, nil
}

// Возвращает список платежей пользователя
func (repo *PostgresPaymentRepo) UserPaymentsList(ctx context.Context, userID string, offset, limit int) ([]models.Payment, error) {
	const op = "PostgresPaymentRepo.UserPaymentsList"
	query := `
		SELECT 
			Payment_id, 
			User_id, 
			Order_id, 
			Amount, 
			Currency,
		    Broker, 
			Operation, 
			Current_status, 
			Created_at
		FROM 
			Transactions
		WHERE 
			User_id = $1
		OFFSET $2 LIMIT $3;`

	rows, err := repo.pool.Query(ctx, query, userID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	paymentList, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Payment, error) {
		var p models.Payment
		err := row.Scan(&p.ID, &p.UserID, &p.OrderID, &p.Amount, &p.Currency,
			&p.Broker, &p.Operation, &p.Status, &p.CreatedAt)
		return p, err
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to collect rows: %w", op, err)
	}

	if len(paymentList) == 0 {
		return nil, ErrPaymentNotFound
	}

	return paymentList, nil
}

// Получает последний статус заказа
func (repo *PostgresPaymentRepo) GetStatus(ctx context.Context, paymentID string) (*models.PaymentStatus, error) {
	const op = "PostgresPaymentRepo.GetStatus"
	query := `
		SELECT 
			s.Status, 
			s.Payment_id, 
			s.Created_at
		FROM 
			TransactionStatus s
		WHERE 
			s.Payment_id = $1
		ORDER BY 
			s.Created_at DESC
		LIMIT 1;`

	var status models.PaymentStatus
	if err := repo.pool.QueryRow(ctx, query, paymentID).
		Scan(&status.Status, &status.PaymentID, &status.CreatedAt); err != nil {

		if err == pgx.ErrNoRows {
			return nil, ErrPaymentStatusNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &status, nil
}

// Обновляет заказ по OrderID
func (repo *PostgresPaymentRepo) UpdateByOrderID(ctx context.Context, transaction models.Payment) error {
	const op = "PostgresPaymentRepo.UpdateByOrderID"
	query := `
		UPDATE Transactions
		SET 
			Payment_id = $1, 
			User_id = $2, 
			Amount = $3,
		    Currency = $4, 
			Broker = $5, 
			Operation = $6
		WHERE 
			Order_id = $7;`

	res, err := repo.pool.Exec(ctx, query,
		transaction.ID, transaction.UserID, transaction.Amount,
		transaction.Currency, transaction.Broker, transaction.Operation,
		transaction.OrderID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

// Удаляет заказ
func (repo *PostgresPaymentRepo) Delete(ctx context.Context, paymentID string) error {
	const op = "PostgresPaymentRepo.Delete"
	query := `DELETE FROM Transactions WHERE Payment_id = $1;`

	res, err := repo.pool.Exec(ctx, query, paymentID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

// Проставляет новый статус заказа
func (repo *PostgresPaymentRepo) MarkStatus(ctx context.Context, paymentID string, status models.StatusType) (err error) {
	const op = "PostgresPaymentRepo.MarkStatus"

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// обновляем текущий статус
	query := `
		UPDATE Transactions
		SET Current_status = $1
		WHERE Payment_id = $2;`

	res, err := tx.Exec(ctx, query, status, paymentID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}

	// вставляем историю по Payment_id
	query = `
		INSERT INTO TransactionStatus(Payment_id, Status)
		VALUES ($1, $2);`
	_, err = tx.Exec(ctx, query, paymentID, status)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit(ctx)
}
