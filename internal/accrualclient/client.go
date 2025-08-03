package accrualclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/etoneja/go-gophermart/internal/models"
)

type AccrualClient struct {
	baseURL string
	client  *http.Client
}

func NewAccrualClient(baseURL string, timeout time.Duration) *AccrualClient {
	return &AccrualClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *AccrualClient) GetOrder(ctx context.Context, orderID string) (*models.AccrualOrderModel, error) {
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
