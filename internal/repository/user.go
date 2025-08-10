package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/etoneja/go-gophermart/internal/errs"
	"github.com/etoneja/go-gophermart/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

type GetUserOptions struct {
	UUID          string
	Login         string
	LockForUpdate bool
}

func (r *UserRepository) CreateUser(ctx context.Context, tx pgx.Tx, user *models.UserModel) error {
	query := `
        INSERT INTO users (
			uuid,
			login,
			hashed_password,
			balance,
			created_at
		)
        VALUES ($1, $2, $3, $4, $5)
    `

	res, err := tx.Exec(
		ctx,
		query,
		user.UUID,
		user.Login,
		user.HashedPassword,
		user.Balance,
		user.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqViolationCode {
			return errs.ErrUserExists
		}
		return err
	}
	if res.RowsAffected() != 1 {
		return errs.ErrNotOnlyOneRowAffected
	}

	return nil
}

func (r *UserRepository) GetUser(ctx context.Context, tx pgx.Tx, opts GetUserOptions) (*models.UserModel, error) {
	query := `
        SELECT
			uuid,
			login,
			hashed_password,
			balance,
			created_at
        FROM users
        WHERE 
    `
	var args []any
	var conditions []string

	if opts.UUID != "" {
		conditions = append(conditions, fmt.Sprintf("uuid = $%d", len(args)+1))
		args = append(args, opts.UUID)
	}

	if opts.Login != "" {
		conditions = append(conditions, fmt.Sprintf("login = $%d", len(args)+1))
		args = append(args, opts.Login)
	}

	if len(conditions) == 0 {
		return nil, fmt.Errorf("no search criteria provided")
	}

	query += strings.Join(conditions, " AND ")

	if opts.LockForUpdate {
		query += " FOR UPDATE"
	}

	var user models.UserModel
	err := tx.QueryRow(ctx, query, args...).Scan(
		&user.UUID,
		&user.Login,
		&user.HashedPassword,
		&user.Balance,
		&user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserBalance(ctx context.Context, tx pgx.Tx, userID string) (*models.BalanceModel, error) {
	query := `
		SELECT 
			u.balance,
			COALESCE(
				(SELECT SUM(amount) 
				FROM transactions 
				WHERE user_id = $1 AND type = $2), 
				0
			) AS withdrawn
		FROM users as u
		WHERE u.uuid = $1;
    `

	var balance models.BalanceModel
	err := tx.QueryRow(ctx, query, userID, models.TransactionTypeWithdraw).Scan(
		&balance.Current,
		&balance.Withdrawn)

	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (r *UserRepository) UpdateUserBalance(ctx context.Context, tx pgx.Tx, transaction *models.TransactionModel) error {
	query := `
		UPDATE users 
		SET balance = balance + $1
		WHERE uuid = $2
	`

	res, err := tx.Exec(
		ctx,
		query,
		transaction.SignedAmount(),
		transaction.UserID)

	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}
	if res.RowsAffected() != 1 {
		return errs.ErrNotOnlyOneRowAffected
	}
	return nil
}

func (r *UserRepository) GetUserWithdrawals(ctx context.Context, tx pgx.Tx, userID string) ([]*models.WithdrawModel, error) {
	query := `
		SELECT
			user_id,
			order_id,
			amount,
			created_at
		FROM transactions as t
		WHERE
			t.user_id = $1
			and t.type = 'withdraw'
		ORDER BY created_at DESC;
    `

	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*models.WithdrawModel
	for rows.Next() {
		var withdraw models.WithdrawModel
		if err := rows.Scan(
			&withdraw.UserID,
			&withdraw.OrderID,
			&withdraw.Sum,
			&withdraw.CreatedAt,
		); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &withdraw)
	}

	return withdrawals, nil
}
