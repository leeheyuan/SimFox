package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		username, err := utils.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("username", username)
		c.Next()
	}
}

func WorkerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		workerName, workerID, tenantID, err := utils.ParseWorkerToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid worker token"})
			return
		}

		if paramValue := c.Param("workerId"); paramValue != "" {
			paramWorkerID, parseErr := strconv.ParseUint(paramValue, 10, 64)
			if parseErr != nil || uint(paramWorkerID) != workerID {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "worker token does not match target worker"})
				return
			}
		}

		c.Set("workerName", workerName)
		c.Set("workerId", workerID)
		c.Set("tenantId", tenantID)
		c.Next()
	}
}
