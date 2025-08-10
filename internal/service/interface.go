package service

import (
	"context"

	"github.com/etoneja/go-gophermart/internal/models"
)

type Servicer interface {
	IsAccrualSytemBusy() bool
	RegisterUser(ctx context.Context, login, password string) (*models.UserModel, string, error)
	LoginUser(ctx context.Context, login, password string) (*models.UserModel, string, error)
	ValidateToken(tokenString string) (string, error)
	GetUserByLogin(ctx context.Context, login string) (*models.UserModel, error)
	GetUserBalance(ctx context.Context, userID string) (*models.BalanceModel, error)
	CreateOrGetOrder(ctx context.Context, order *models.OrderModel) (*models.OrderModel, error)
	GetOrdersForUser(ctx context.Context, user *models.UserModel) (models.OrderModelList, error)
	GetOrdersToSync(ctx context.Context, limit int) (models.OrderModelList, error)
	GetOrder(ctx context.Context, orderID string) (*models.OrderModel, error)
	GetUserWithdrawals(ctx context.Context, userID string) (models.WithdrawModelList, error)
	CreateWithdraw(ctx context.Context, withdraw *models.WithdrawModel) error
	SyncOrder(ctx context.Context, orderID string) error
}
