package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit returns a middleware that enforces a fixed-window rate limit of
// maxRequests per window, keyed by client IP. State is kept in process memory
// so the limit is per instance; a janitor goroutine evicts idle buckets every
// 2*window to bound memory.
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	return RateLimitFunc(window, func(*gin.Context) int {
		return maxRequests
	})
}

// RateLimitFunc is a dynamic variant of [RateLimit] that resolves the allowed
// request count for every incoming request. This allows administrators to tune
// the limit at runtime without restarting the server.
func RateLimitFunc(window time.Duration, resolveMax func(*gin.Context) int) gin.HandlerFunc {
	type bucket struct {
		count     int
		lastReset time.Time
	}

	var mu sync.Mutex
	buckets := make(map[string]*bucket)

	go func() {
		for {
			time.Sleep(window * 2)
			mu.Lock()
			now := time.Now()
			for ip, b := range buckets {
				if now.Sub(b.lastReset) > window*2 {
					delete(buckets, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		maxRequests := resolveMax(c)
		if maxRequests < 1 {
			c.Next()
			return
		}

		ip := c.ClientIP()
		mu.Lock()

		b, exists := buckets[ip]
		now := time.Now()

		if !exists || now.Sub(b.lastReset) >= window {
			buckets[ip] = &bucket{count: 1, lastReset: now}
			mu.Unlock()
			c.Next()
			return
		}

		if b.count >= maxRequests {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    42900,
				"message": "too many requests, please try again later",
				"data":    nil,
			})
			c.Abort()
			return
		}

		b.count++
		mu.Unlock()
		c.Next()
	}
}
