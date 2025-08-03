package handlers

import (
	"errors"

	"github.com/etoneja/go-gophermart/internal/models"
	"github.com/gin-gonic/gin"
)

func GetCurrentUser(c *gin.Context) (*models.UserModel, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, errors.New("user not authenticated")
	}

	usr, ok := user.(*models.UserModel)
	if !ok {
		return nil, errors.New("invalid user type")
	}

	return usr, nil
}
