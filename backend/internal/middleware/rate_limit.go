package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter 限流器：优先使用 Redis 固定窗口计数，Redis 不可用时回退进程内计数。
type RateLimiter struct {
	rdb       *redis.Client
	limit     int
	window    time.Duration
	mu        sync.Mutex
	memCounts map[string]*memBucket
}

type memBucket struct {
	count int
	reset time.Time
}

// NewRateLimiter 构造限流器。
func NewRateLimiter(rdb *redis.Client, limitPerMin int) *RateLimiter {
	if limitPerMin <= 0 {
		limitPerMin = 120
	}
	return &RateLimiter{
		rdb:       rdb,
		limit:     limitPerMin,
		window:    time.Minute,
		memCounts: make(map[string]*memBucket),
	}
}

// Middleware 生成 Gin 限流中间件。
func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "rl:" + c.ClientIP()
		allowed, err := r.allowRedis(c.Request.Context(), key)
		if err != nil {
			allowed = r.allowMem(key)
		}
		if !allowed {
			util.Fail(c, util.NewAppError(constants.CodeRateLimited, constants.MsgErrRateLimited, nil))
			c.Abort()
			return
		}
		c.Next()
	}
}

func (r *RateLimiter) allowRedis(ctx context.Context, key string) (bool, error) {
	now := time.Now().Unix()
	windowKey := fmt.Sprintf("%s:%d", key, now/60)
	pipe := r.rdb.Pipeline()
	incr := pipe.Incr(ctx, windowKey)
	pipe.Expire(ctx, windowKey, 2*r.window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	count, err := incr.Result()
	if err != nil {
		return false, err
	}
	return count <= int64(r.limit), nil
}

func (r *RateLimiter) allowMem(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	b, ok := r.memCounts[key]
	if !ok || now.After(b.reset) {
		r.memCounts[key] = &memBucket{count: 1, reset: now.Add(r.window)}
		return true
	}
	b.count++
	return b.count <= r.limit
}
