package accrualclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/etoneja/go-gophermart/internal/errs"
	"github.com/etoneja/go-gophermart/internal/models"
)

var defaultRetryAfterOnRateLimit = 30 * time.Second

type AccrualClient struct {
	baseURL     string
	client      *http.Client
	rateLimiter *rateLimiter
}

func NewAccrualClient(baseURL string, timeout time.Duration) *AccrualClient {
	return &AccrualClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
		rateLimiter: NewRateLimiter(),
	}
}

func (c *AccrualClient) IsRateLimited() bool {
	return c.rateLimiter.isBlocked()
}

func (c *AccrualClient) GetOrder(ctx context.Context, orderID string) (*models.AccrualOrderModel, error) {
	if c.IsRateLimited() {
		return nil, errs.ErrRateLimit
	}

	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		c.rateLimiter.blockFor(retryAfter)
		return nil, errs.ErrRateLimit
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var accrualOrderResponse models.AccrualOrderResponse
	if err := json.Unmarshal(body, &accrualOrderResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return accrualOrderResponse.ToModel(), nil
}

func parseRetryAfter(header string) time.Duration {
	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second
	}

	if t, err := http.ParseTime(header); err == nil {
		return time.Until(t)
	}

	return defaultRetryAfterOnRateLimit
}
