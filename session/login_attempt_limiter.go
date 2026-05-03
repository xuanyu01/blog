/*
login_attempt_limiter.go 提供基于 Redis 的登录失败限流实现。*/
package session

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoginAttemptLimiter 负责记录和校验登录失败次数。
type LoginAttemptLimiter struct {
	client        *redis.Client
	ctx           context.Context
	maxAttempts   int
	window        time.Duration
	blockDuration time.Duration
}

// NewLoginAttemptLimiter 创建登录限流器。
func NewLoginAttemptLimiter(client *redis.Client, maxAttempts int, window, blockDuration time.Duration) *LoginAttemptLimiter {
	return &LoginAttemptLimiter{
		client:        client,
		ctx:           context.Background(),
		maxAttempts:   maxAttempts,
		window:        window,
		blockDuration: blockDuration,
	}
}

// Check 返回当前键是否仍处于封禁期内。
func (l *LoginAttemptLimiter) Check(key string) (time.Duration, error) {
	if l == nil || key == "" {
		return 0, nil
	}

	retryAfter, err := l.client.TTL(l.ctx, l.blockKey(key)).Result()
	if err != nil {
		return 0, err
	}
	if retryAfter < 0 {
		return 0, nil
	}
	return retryAfter, nil
}

// RegisterFailure 记录一次失败登录，并在达到阈值后返回封禁时间。
func (l *LoginAttemptLimiter) RegisterFailure(key string) (time.Duration, error) {
	if l == nil || key == "" {
		return 0, nil
	}

	if retryAfter, err := l.Check(key); err != nil || retryAfter > 0 {
		return retryAfter, err
	}

	count, err := l.client.Incr(l.ctx, l.countKey(key)).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := l.client.Expire(l.ctx, l.countKey(key), l.window).Err(); err != nil {
			return 0, err
		}
	}

	if int(count) < l.maxAttempts {
		return 0, nil
	}

	pipe := l.client.TxPipeline()
	pipe.Set(l.ctx, l.blockKey(key), "1", l.blockDuration)
	pipe.Del(l.ctx, l.countKey(key))
	if _, err := pipe.Exec(l.ctx); err != nil {
		return 0, err
	}

	return l.blockDuration, nil
}

// Reset 清空指定键的失败次数和封禁状态。
func (l *LoginAttemptLimiter) Reset(key string) error {
	if l == nil || key == "" {
		return nil
	}

	return l.client.Del(l.ctx, l.countKey(key), l.blockKey(key)).Err()
}

// countKey 返回失败次数统计键。
func (l *LoginAttemptLimiter) countKey(key string) string {
	return fmt.Sprintf("login_limit:count:%s", key)
}

// blockKey 返回封禁状态键。
func (l *LoginAttemptLimiter) blockKey(key string) string {
	return fmt.Sprintf("login_limit:block:%s", key)
}

