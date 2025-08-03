package handlers

import (
	"errors"
	"net/http"

	"github.com/etoneja/go-gophermart/internal/errs"
	"github.com/etoneja/go-gophermart/internal/models"
	"github.com/etoneja/go-gophermart/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (h *Handlers) GetBalanceHandler(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	balance, err := h.svc.GetUserBalance(c.Request.Context(), user.UUID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get balance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get balance"})
		return
	}

	c.JSON(http.StatusOK, balance.ToResponse())
}

func (h *Handlers) GetWithdrawalsHandler(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	withdrawals, err := h.svc.GetUserWithdrawals(c.Request.Context(), user.UUID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get withdrawals")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get withdrawals"})
		return
	}

	if len(withdrawals) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"error": "no withdrawals for current user"})
		return
	}

	c.JSON(http.StatusOK, withdrawals.ToResponse())
}

func (h *Handlers) CreateWithdrawHandler(c *gin.Context) {
	var req models.WithdrawRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	isOrderIDValid, err := utils.LuhnCheck(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad orderID format"})
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

	withdraw := req.ToModel(user.UUID)

	err = h.svc.CreateWithdraw(c.Request.Context(), withdraw)
	if err != nil {
		if errors.Is(err, errs.ErrInsufficientFunds) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient funds"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get balance"})
		return
	}

	c.JSON(http.StatusOK, withdraw.ToResponse())
}
