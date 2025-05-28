package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.RWMutex
)

func RateLimiter(requestsPerSecond float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.RLock()
		limiter, exists := limiters[ip]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), int(requestsPerSecond*2))
			limiters[ip] = limiter
			mu.Unlock()
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
