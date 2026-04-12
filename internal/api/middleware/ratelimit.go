package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit 基于令牌桶的简单速率限制中间件。
// 每个 IP 地址独立限流，不依赖 Redis。
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	type bucket struct {
		tokens    int
		lastReset time.Time
	}

	var mu sync.Mutex
	buckets := make(map[string]*bucket)

	// 定期清理过期条目
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
		ip := c.ClientIP()
		mu.Lock()

		b, exists := buckets[ip]
		now := time.Now()

		if !exists || now.Sub(b.lastReset) >= window {
			buckets[ip] = &bucket{tokens: maxRequests - 1, lastReset: now}
			mu.Unlock()
			c.Next()
			return
		}

		if b.tokens <= 0 {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    42900,
				"message": "too many requests, please try again later",
				"data":    nil,
			})
			c.Abort()
			return
		}

		b.tokens--
		mu.Unlock()
		c.Next()
	}
}
