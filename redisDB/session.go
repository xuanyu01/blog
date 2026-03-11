package redisDB

import (
	"time"

	"github.com/google/uuid"
)

// Session过期时间
const SessionExpire = 1 * time.Hour

// 创建session
func CreateSession(userID string) (string, error) {

	sessionID := uuid.New().String()

	key := "session:" + sessionID

	err := RedisClient.Set(ctx, key, userID, SessionExpire).Err()
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

// 获取session
func GetSession(sessionID string) (string, error) {

	key := "session:" + sessionID

	userID, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return userID, nil
}

// 删除session
func DeleteSession(sessionID string) error {

	key := "session:" + sessionID

	return RedisClient.Del(ctx, key).Err()
}
