package accrualclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/etoneja/go-gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccrualClient_GetOrder(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		wantResult     *models.AccrualOrderModel
		wantError      string
	}{
		{
			name:           "successful response",
			responseStatus: http.StatusOK,
			responseBody:   `{"order":"123","status":"PROCESSED","accrual":100.5}`,
			wantResult: &models.AccrualOrderModel{
				ID:     "123",
				Status: models.AccrualOrderStatusProcessed,
				Accrual: func() *int64 {
					v := int64(10050)
					return &v
				}(),
			},
		},
		{
			name:           "order not found",
			responseStatus: http.StatusNoContent,
			wantError:      "unexpected status code: 204",
		},
		{
			name:           "invalid json",
			responseStatus: http.StatusOK,
			responseBody:   `{"order":123}`,
			wantError:      "failed to unmarshal response",
		},
		{
			name:           "server error",
			responseStatus: http.StatusInternalServerError,
			wantError:      "unexpected status code: 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/orders/123", r.URL.Path)
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != "" {
					_, _ = w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			client := NewAccrualClient(server.URL, 1*time.Second)

			result, err := client.GetOrder(context.Background(), "123")

			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

func TestAccrualClient_RequestErrors(t *testing.T) {
	t.Run("request timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL, 100*time.Millisecond)

		_, err := client.GetOrder(context.Background(), "123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request failed")
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Second)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL, 5*time.Second)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.GetOrder(ctx, "123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
