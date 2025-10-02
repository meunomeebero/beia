package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis *redis.Client
	ctx   context.Context
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis: redisClient,
		ctx:   context.Background(),
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for OPTIONS requests (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		ip := c.ClientIP()

		blacklistKey := fmt.Sprintf("blacklist:%s", ip)
		isBlacklisted, err := rl.redis.Exists(rl.ctx, blacklistKey).Result()
		if err == nil && isBlacklisted > 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "IP banned for violating rate limits",
			})
			c.Abort()
			return
		}

		dailyKey := fmt.Sprintf("daily:%s:%s", ip, time.Now().Format("2006-01-02"))
		dailyCount, err := rl.redis.Get(rl.ctx, dailyKey).Int()
		if err != nil && err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}

		if dailyCount >= 100 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Daily limit of 100 requests exceeded",
			})
			c.Abort()
			return
		}

		minuteKey := fmt.Sprintf("ratelimit:%s", ip)
		requestCount, err := rl.redis.Get(rl.ctx, minuteKey).Int()
		if err != nil && err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}

		if requestCount >= 10 {
			rl.redis.Set(rl.ctx, blacklistKey, "1", 7*24*time.Hour)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Rate limit exceeded. IP banned for 1 week",
			})
			c.Abort()
			return
		}

		pipe := rl.redis.Pipeline()

		pipe.Incr(rl.ctx, minuteKey)
		pipe.Expire(rl.ctx, minuteKey, time.Minute)

		pipe.Incr(rl.ctx, dailyKey)
		pipe.ExpireAt(rl.ctx, dailyKey, time.Now().Add(24*time.Hour))

		_, err = pipe.Exec(rl.ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
