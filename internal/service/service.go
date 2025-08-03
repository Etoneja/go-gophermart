package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/etoneja/go-gophermart/internal/accrualclient"
	"github.com/etoneja/go-gophermart/internal/config"
	"github.com/etoneja/go-gophermart/internal/db"
	"github.com/etoneja/go-gophermart/internal/errs"
	"github.com/etoneja/go-gophermart/internal/models"
	"github.com/etoneja/go-gophermart/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Service struct {
	cfg           *config.Config
	dbPool        *pgxpool.Pool
	logger        zerolog.Logger
	accrualClient accrualclient.AccrualClienter
	repos         *repository.Repositories
}

func NewService(cfg *config.Config, dbPool *pgxpool.Pool, logger zerolog.Logger) *Service {

	accrualClient := accrualclient.NewAccrualClient(cfg.AccrualSystemAddress, 10*time.Second)
	repos := repository.NewRepositories()

	return &Service{
		cfg:           cfg,
		dbPool:        dbPool,
		accrualClient: accrualClient,
		repos:         repos,
	}
}

func (s *Service) RegisterUser(ctx context.Context, login, password string) (*models.UserModel, string, error) {
	hashedPassword, err := models.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	var token string
	user := &models.UserModel{
		UUID:           uuid.NewString(),
		Login:          login,
		HashedPassword: hashedPassword,
		Balance:        0,
		CreatedAt:      time.Now(),
	}

	err = db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {
		tx := db.GetTxFromContext(txCtx)
		err := s.repos.UserRepo.CreateUser(txCtx, tx, user)
		if err != nil {
			return err
		}
		token, err = s.generateJWTToken(user.Login)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	return user, token, nil
}

func (s *Service) LoginUser(ctx context.Context, login, password string) (*models.UserModel, string, error) {
	var user *models.UserModel
	var token string

	err := db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {
		tx := db.GetTxFromContext(txCtx)

		var err error
		getUserOps := repository.GetUserOptions{Login: login}
		user, err = s.repos.UserRepo.GetUser(txCtx, tx, getUserOps)
		if err != nil {
			return errs.ErrInvalidCredentials
		}

		if !models.CheckPasswordHash(password, user.HashedPassword) {
			return errs.ErrInvalidCredentials
		}

		token, err = s.generateJWTToken(user.Login)
		return err
	})

	if err != nil {
		return nil, "", fmt.Errorf("login failed: %w", err)
	}

	return user, token, nil
}

func (s *Service) generateJWTToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) ValidateToken(tokenString string) (string, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return "", errors.New("invalid token claims")
		}
		return username, nil
	}
	return "", errors.New("invalid token")
}

func (s *Service) GetUserByLogin(ctx context.Context, login string) (*models.UserModel, error) {
	var user *models.UserModel
	err := db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {
		tx := db.GetTxFromContext(txCtx)

		var err error
		opts := repository.GetUserOptions{Login: login}
		user, err = s.repos.UserRepo.GetUser(txCtx, tx, opts)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) GetUserBalance(ctx context.Context, userID string) (*models.BalanceModel, error) {
	var balance *models.BalanceModel
	err := db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {
		tx := db.GetTxFromContext(txCtx)

		var err error
		balance, err = s.repos.UserRepo.GetUserBalance(txCtx, tx, userID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return balance, nil

}

func (s *Service) CreateOrGetOrder(ctx context.Context, order *models.OrderModel) (*models.OrderModel, error) {
	err := db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {
		tx := db.GetTxFromContext(txCtx)
		return s.repos.OrderRepo.CreateOrder(ctx, tx, order)
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *Service) GetOrdersForUser(ctx context.Context, user *models.UserModel) (models.OrderModelList, error) {
	var orders models.OrderModelList
	err := db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {
		tx := db.GetTxFromContext(txCtx)
		var err error
		orders, err = s.repos.OrderRepo.GetOrdersForUserID(ctx, tx, user.UUID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *Service) GetOrdersToSync(ctx context.Context, limit int) (models.OrderModelList, error) {
	var orders models.OrderModelList
	err := db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {
		tx := db.GetTxFromContext(txCtx)
		var err error
		orders, err = s.repos.OrderRepo.GetOrdersToSync(ctx, tx, limit)
		return err
	})
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*models.OrderModel, error) {
	var order *models.OrderModel
	err := db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {
		tx := db.GetTxFromContext(txCtx)

		var err error
		opts := repository.GetOrderOptions{ID: orderID}
		order, err = s.repos.OrderRepo.GetOrder(txCtx, tx, opts)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return order, nil

}

func (s *Service) GetUserWithdrawals(ctx context.Context, userID string) (models.WithdrawModelList, error) {
	var withdrawals models.WithdrawModelList
	err := db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {
		tx := db.GetTxFromContext(txCtx)
		var err error
		withdrawals, err = s.repos.UserRepo.GetUserWithdrawals(ctx, tx, userID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (s *Service) CreateWithdraw(ctx context.Context, withdraw *models.WithdrawModel) error {
	err := db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {
		if withdraw.Sum <= 0 {
			return errors.New("withdraw Sum should be positive")
		}

		tx := db.GetTxFromContext(txCtx)

		opts := repository.GetUserOptions{
			UUID:          withdraw.UserID,
			LockForUpdate: true,
		}
		user, err := s.repos.UserRepo.GetUser(txCtx, tx, opts)
		if err != nil {
			return fmt.Errorf("can't get user: %w", err)
		}

		if user.Balance < withdraw.Sum {
			return errs.ErrInsufficientFunds
		}

		transaction := &models.TransactionModel{
			UUID:      uuid.NewString(),
			UserID:    withdraw.UserID,
			OrderID:   withdraw.OrderID,
			Type:      models.TransactionTypeWithdraw,
			Amount:    withdraw.Sum,
			CreatedAt: withdraw.CreatedAt,
		}

		err = s.repos.TransactionRepo.CreateTransaction(ctx, tx, transaction)
		if err != nil {
			return fmt.Errorf("can't create transaction: %w", err)
		}

		err = s.repos.UserRepo.UpdateUserBalance(txCtx, tx, transaction)
		if err != nil {
			return fmt.Errorf("can't update user balance: %w", err)
		}

		return nil
	})
	return err
}

func (s *Service) SyncOrder(ctx context.Context, orderID string) error {
	err := db.WithTx(ctx, s.dbPool, func(txCtx context.Context) error {

		tx := db.GetTxFromContext(txCtx)

		getOrderOpts := repository.GetOrderOptions{
			ID:            orderID,
			LockForUpdate: true,
			SkipLocked:    true,
		}
		order, err := s.repos.OrderRepo.GetOrder(txCtx, tx, getOrderOpts)
		if err != nil {
			if errors.Is(err, errs.ErrNoRows) {
				s.logger.Warn().
					Str("orderID", orderID).
					Err(err).
					Msg("can't get order for update, skipping...")
				return nil
			}
			return fmt.Errorf("failed to get order from db: %w", err)
		}
		if order.IsTerminated() {
			s.logger.Info().
				Str("orderID", orderID).
				Err(err).
				Msg("order already processed, skipping...")
			return nil
		}

		accrualOrder, err := s.accrualClient.GetOrder(txCtx, orderID)
		if err != nil {
			return fmt.Errorf("failed to get order from accrual system: %w", err)
		}

		newOrderStatus, err := models.ConvertAccrualOrderStatusToOrderStatus(accrualOrder.Status)
		if err != nil {
			return err
		}

		order.Status = newOrderStatus
		order.UpdatedAt = time.Now()
		order.Accrual = accrualOrder.Accrual

		err = s.repos.OrderRepo.UpdateOrder(txCtx, tx, order)
		if err != nil {
			return fmt.Errorf("can't update order: %w", err)
		}

		if accrualOrder.Status == models.AccrualOrderStatusProcessed {
			getUserOps := repository.GetUserOptions{
				UUID:          order.UserID,
				LockForUpdate: true,
			}
			_, err = s.repos.UserRepo.GetUser(txCtx, tx, getUserOps)
			if err != nil {
				return fmt.Errorf("can't get user: %w", err)
			}

			transaction := &models.TransactionModel{
				UUID:      uuid.NewString(),
				UserID:    order.UserID,
				OrderID:   order.ID,
				Type:      models.TransactionTypeAccrual,
				Amount:    *order.Accrual,
				CreatedAt: time.Now(),
			}

			err = s.repos.TransactionRepo.CreateTransaction(ctx, tx, transaction)
			if err != nil {
				return fmt.Errorf("can't create transaction: %w", err)
			}

			err = s.repos.UserRepo.UpdateUserBalance(txCtx, tx, transaction)
			if err != nil {
				return fmt.Errorf("can't update user balance: %w", err)
			}

		}
		s.logger.Info().
			Str("orderID", orderID).
			Msg("order processed sucessfully")
		return nil
	})

	return err
}
