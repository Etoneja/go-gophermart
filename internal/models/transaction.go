package models

import "time"

type TransactionType string

const (
	TransactionTypeAccrual  TransactionType = "accrual"
	TransactionTypeWithdraw TransactionType = "withdraw"
)

type TransactionModel struct {
	UUID      string          `json:"-"`
	UserID    string          `json:"-"`
	OrderID   string          `json:"-"`
	Type      TransactionType `json:"-"`
	Amount    int64           `json:"-"`
	CreatedAt time.Time       `json:"-"`
}

func (t *TransactionModel) SignedAmount() int64 {
	if t.Type == TransactionTypeWithdraw {
		return -t.Amount
	}
	return t.Amount
}
