package repository

import (
	"context"

	"github.com/etoneja/go-gophermart/internal/models"
	"github.com/jackc/pgx/v5"
)

type TransactionRepository struct{}

func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{}
}

func (r *TransactionRepository) CreateTransaction(ctx context.Context, tx pgx.Tx, transaction *models.TransactionModel) error {
	query := `
		INSERT INTO transactions (
			uuid,
			user_id,
			order_id,
			type,
			amount,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.Exec(
		ctx,
		query,
		transaction.UUID,
		transaction.UserID,
		transaction.OrderID,
		transaction.Type,
		transaction.Amount,
		transaction.CreatedAt)

	return err
}
