package accrualclient

import (
	"context"

	"github.com/etoneja/go-gophermart/internal/models"
)

type AccrualClienter interface {
	GetOrder(ctx context.Context, orderID string) (*models.AccrualOrderModel, error)
}
