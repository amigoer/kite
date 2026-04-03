package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipRecord struct {
	count    int
	firstAt  time.Time
}

// RateLimiter IP 级别频率限制器
type RateLimiter struct {
	mu       sync.RWMutex
	records  map[string]*ipRecord
	maxCount int
	window   time.Duration
}

// NewRateLimiter 创建频率限制器
func NewRateLimiter(maxCount int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		records:  make(map[string]*ipRecord),
		maxCount: maxCount,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

// cleanup 定期清理过期记录
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, rec := range rl.records {
			if now.Sub(rec.firstAt) > rl.window {
				delete(rl.records, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware 返回 Gin 中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		now := time.Now()
		rec, exists := rl.records[ip]

		if !exists || now.Sub(rec.firstAt) > rl.window {
			rl.records[ip] = &ipRecord{count: 1, firstAt: now}
			rl.mu.Unlock()
			c.Next()
			return
		}

		rec.count++
		if rec.count > rl.maxCount {
			rl.mu.Unlock()
			Error(c, http.StatusTooManyRequests, 429, "too many requests, please try again later")
			c.Abort()
			return
		}

		rl.mu.Unlock()
		c.Next()
	}
}
