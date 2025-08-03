package repository

import (
	"context"
	"errors"

	"github.com/etoneja/go-gophermart/internal/errs"
	"github.com/etoneja/go-gophermart/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrderRepository struct{}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{}
}

type GetOrderOptions struct {
	ID            string
	LockForUpdate bool
	SkipLocked    bool
}

func (r *OrderRepository) GetOrder(ctx context.Context, tx pgx.Tx, opts GetOrderOptions) (*models.OrderModel, error) {
	query := `
		SELECT
			id,
			user_id,
			status,
			accrual,
			created_at,
			updated_at
		FROM orders
		WHERE id = $1
	`

	if opts.LockForUpdate {
		query += " FOR UPDATE"

		if opts.SkipLocked {
			query += " SKIP LOCKED"
		}
	}

	var order models.OrderModel
	err := tx.QueryRow(ctx, query, opts.ID).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.CreatedAt,
		&order.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNoRows
		}
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, tx pgx.Tx, order *models.OrderModel) error {
	query := `
		INSERT INTO orders (
			id,
			user_id,
			status,
			accrual,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := tx.QueryRow(
		ctx,
		query,
		order.ID,
		order.UserID,
		order.Status,
		order.Accrual,
		order.CreatedAt,
		order.UpdatedAt).Scan(&order.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqViolationCode {
			return errs.ErrOrderExists
		}
		return err
	}
	return nil
}

func (r *OrderRepository) GetOrdersToSync(ctx context.Context, tx pgx.Tx, limit int) (models.OrderModelList, error) {
	query := `
		SELECT
			id,
			user_id,
			status,
			accrual,
			created_at,
			updated_at
		FROM orders
		WHERE status != all($1)
		ORDER BY updated_at asc
		FOR UPDATE SKIP LOCKED
		LIMIT $2
	`

	rows, err := tx.Query(
		ctx,
		query,
		models.TerminatedOrderStatuses,
		limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.fetchOrders(rows)
}

func (r *OrderRepository) GetOrdersForUserID(ctx context.Context, tx pgx.Tx, userID string) (models.OrderModelList, error) {
	query := `
		SELECT
			id,
			user_id,
			status,
			accrual,
			created_at,
			updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at asc
	`

	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.fetchOrders(rows)

}

func (r *OrderRepository) fetchOrders(rows pgx.Rows) (models.OrderModelList, error) {
	var orders models.OrderModelList
	for rows.Next() {
		var order models.OrderModel
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, tx pgx.Tx, order *models.OrderModel) error {
	query := `
		UPDATE orders
		SET
			status = $1,
			accrual = $2,
			updated_at = $3
		WHERE id = $4
	`

	res, err := tx.Exec(
		ctx,
		query,
		order.Status,
		order.Accrual,
		order.UpdatedAt,
		order.ID)
	if err != nil {
		return err
	}
	if res.RowsAffected() != 1 {
		return errs.ErrNotOnlyOneRowAffected
	}
	return nil
}
