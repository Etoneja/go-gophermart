package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/etoneja/go-gophermart/internal/errs"
	"github.com/etoneja/go-gophermart/internal/models"
	"github.com/etoneja/go-gophermart/internal/utils"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) CreateOrderHandler(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	orderID := string(body)

	isOrderIDValid, err := utils.LuhnCheck(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if !isOrderIDValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "order id not in luhn format"})
		return
	}

	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	order := &models.OrderModel{
		ID:        orderID,
		UserID:    user.UUID,
		Status:    models.OrderStatusNew,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdOrder, err := h.svc.CreateOrGetOrder(c.Request.Context(), order)
	if err != nil {
		if errors.Is(err, errs.ErrOrderExists) {
			oldOrder, err := h.svc.GetOrder(c.Request.Context(), order.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "err fetching old order on conflict"})
				return
			}
			if order.UserID == oldOrder.UserID {
				c.JSON(http.StatusOK, oldOrder.ToResponse())
				return
			}
			c.JSON(http.StatusConflict, gin.H{"error": "order already created by another user"})
			return
		}

		h.logger.Error().Err(err).Msg("Failed to create order")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	c.JSON(http.StatusAccepted, createdOrder.ToResponse())
}

func (h *Handlers) GetOrdersHandler(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	orders, err := h.svc.GetOrdersForUser(c.Request.Context(), user)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get orders")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get orders"})
		return
	}

	if len(orders) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"error": "no orders for current user"})
		return
	}

	c.JSON(http.StatusOK, orders.ToResponse())
}
