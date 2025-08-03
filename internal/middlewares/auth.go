package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Middlewares) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		login, err := m.svc.ValidateToken(tokenString)
		if err != nil {
			m.logger.Warn().Err(err).Msg("Invalid token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		user, err := m.svc.GetUserByLogin(c, login)
		if err != nil {
			m.logger.Warn().Err(err).Msg("Can't fetch user")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "can't fetch user"})
			return
		}

		c.Set("user", user)

		c.Next()
	}
}
