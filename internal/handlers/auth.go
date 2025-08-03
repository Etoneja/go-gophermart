package handlers

import (
	"errors"
	"net/http"

	"github.com/etoneja/go-gophermart/internal/errs"
	"github.com/gin-gonic/gin"
)

const AuthorizationHeader = "Authorization"

type RegisterRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=25"`
	Password string `json:"password" binding:"required,min=3"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handlers) RegisterUserHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn().Err(err).Msg("Invalid registration payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, token, err := h.svc.RegisterUser(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "login already registered"})
			return
		}
		h.logger.Error().Err(err).Msg("Failed to register user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.Header(AuthorizationHeader, token)
	c.Status(http.StatusOK)
}

func (h *Handlers) LoginUserHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn().Err(err).Msg("Invalid login payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, token, err := h.svc.LoginUser(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		h.logger.Warn().Err(err).Str("username", req.Login).Msg("Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.Header(AuthorizationHeader, token)
	c.Status(http.StatusOK)
}
