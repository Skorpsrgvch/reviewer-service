package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет Bearer токен администратора
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "Bearer admin" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "admin token required",
				},
			})
			return
		}

		// Если токен валиден, передаем управление следующему обработчику
		c.Next()
	}
}
