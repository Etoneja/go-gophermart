package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/etoneja/go-gophermart/internal/config"
	"github.com/etoneja/go-gophermart/internal/db"
	"github.com/etoneja/go-gophermart/internal/handlers"
	"github.com/etoneja/go-gophermart/internal/middlewares"
	"github.com/etoneja/go-gophermart/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type APIApp struct {
	Config *config.Config
	DB     *pgxpool.Pool
	Router *gin.Engine
	Server *http.Server
}

func NewAPIApp(ctx context.Context, logger zerolog.Logger) (*APIApp, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	dbPool, err := db.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	migrator := db.NewMigrator(dbPool, logger)
	if err := migrator.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	svc := service.NewService(cfg, dbPool, logger)

	mws := middlewares.NewMiddlewares(svc, logger)

	router := gin.New()
	router.Use(gin.Recovery())

	if cfg.Debug {
		router.Use(mws.DebugLoggingMiddleware())
	} else {
		router.Use(mws.LoggingMiddleware())
	}

	hs := handlers.NewHandlers(svc, logger)

	apiGroup := router.Group("/api/user")
	{
		apiGroup.POST("/register", hs.RegisterUserHandler)
		apiGroup.POST("/login", hs.LoginUserHandler)

		privateGroup := apiGroup.Group("")
		privateGroup.Use(mws.AuthMiddleware())
		{
			privateGroup.POST("/orders", hs.CreateOrderHandler)
			privateGroup.GET("/orders", hs.GetOrdersHandler)
			privateGroup.GET("/balance", hs.GetBalanceHandler)
			privateGroup.POST("/balance/withdraw", hs.CreateWithdrawHandler)
			privateGroup.GET("/withdrawals", hs.GetWithdrawalsHandler)
		}
	}

	return &APIApp{
		Config: cfg,
		DB:     dbPool,
		Router: router,
	}, nil
}

func (a *APIApp) Run() error {
	a.Server = &http.Server{
		Addr:    a.Config.ServerAddress,
		Handler: a.Router,
	}

	if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (a *APIApp) Shutdown(ctx context.Context) error {
	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}
	a.DB.Close()
	return nil
}
