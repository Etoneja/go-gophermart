package processor

import (
	"context"
	"sync"
	"time"

	"github.com/etoneja/go-gophermart/internal/config"
	"github.com/etoneja/go-gophermart/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type OrderProcessor struct {
	cfg    *config.Config
	svc    service.Servicer
	wg     sync.WaitGroup
	logger zerolog.Logger
}

func NewOrderProcessor(cfg *config.Config, dbPool *pgxpool.Pool, logger zerolog.Logger) *OrderProcessor {

	svc := service.NewService(cfg, dbPool, logger)

	return &OrderProcessor{
		cfg:    cfg,
		svc:    svc,
		logger: logger,
	}
}

func (p *OrderProcessor) Run(ctx context.Context) {
	p.wg.Add(1)
	defer p.wg.Done()

	ticker := time.NewTicker(p.cfg.WorkerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info().Msg("Order processor stopped")
			return
		case <-ticker.C:
			p.processOrders(ctx)
		}
	}
}

func (p *OrderProcessor) processOrders(ctx context.Context) {
	orders, err := p.svc.GetOrdersToSync(ctx, p.cfg.WorkerPoolSize)
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to get pending orders")
		return
	}

	if len(orders) == 0 {
		p.logger.Info().Msg("No orders to process")
		return
	}

	p.logger.Info().Int("count", len(orders)).Msg("Processing orders")

	var (
		batchWg sync.WaitGroup
		sem     = make(chan struct{}, p.cfg.WorkerPoolSize)
	)

	for _, order := range orders {
		select {
		case <-ctx.Done():
			batchWg.Wait()
			return
		case sem <- struct{}{}:
			batchWg.Add(1)

			go func(orderID string) {
				defer batchWg.Done()
				defer func() { <-sem }()

				if err := p.svc.SyncOrder(ctx, orderID); err != nil {
					p.logger.Error().Err(err).Str("orderID", orderID).Msg("Order processing failed")
				}
			}(order.ID)
		}
	}

	batchWg.Wait()

}

func (p *OrderProcessor) Stop() {
	p.wg.Wait()
}
