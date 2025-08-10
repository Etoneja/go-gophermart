package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ctxKey string

const txKey ctxKey = "tx"

func BeginTx(ctx context.Context, db *pgxpool.Pool) (context.Context, pgx.Tx, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	return context.WithValue(ctx, txKey, tx), tx, nil
}

func GetTxFromContext(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return nil
}

func WithTx(ctx context.Context, db *pgxpool.Pool, fn func(context.Context) error) error {
	txCtx, tx, err := BeginTx(ctx, db)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(txCtx)
			panic(p)
		} else if err != nil {
			tx.Rollback(txCtx)
		} else {
			err = tx.Commit(txCtx)
		}
	}()

	err = fn(txCtx)
	return err
}
