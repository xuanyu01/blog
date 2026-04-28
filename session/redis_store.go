/*
redis_store.go 提供基于 Redis 的会话存储实现。
*/
package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisStore 是会话存储接口的 Redis 实现
// 它负责把用户标识映射为带过期时间的会话键
type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStore 创建 Redis 会话存储实例
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
		ctx:    context.Background(),
	}
}

// Create 创建新的会话并返回会话标识
func (s *RedisStore) Create(userID string) (string, error) {
	sessionID := uuid.New().String()
	key := SessionPrefix + sessionID

	// 统一添加前缀，便于后续排查和区分不同类型的 Redis 键
	if err := s.client.Set(s.ctx, key, userID, Expire).Err(); err != nil {
		return "", err
	}

	return sessionID, nil
}

// Get 根据会话标识读取用户标识
func (s *RedisStore) Get(sessionID string) (string, error) {
	return s.client.Get(s.ctx, SessionPrefix+sessionID).Result()
}

// Update 更新会话对应的用户标识并刷新过期时间
func (s *RedisStore) Update(sessionID, userID string) error {
	return s.client.Set(s.ctx, SessionPrefix+sessionID, userID, Expire).Err()
}

// Delete 删除指定会话
func (s *RedisStore) Delete(sessionID string) error {
	return s.client.Del(s.ctx, SessionPrefix+sessionID).Err()
}
