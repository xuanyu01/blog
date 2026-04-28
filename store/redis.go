/*
redis.go 负责初始化 Redis 客户端连接。
*/
package store

import (
	"blog/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// NewRedis 创建并返回可用的 Redis 客户端
func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		Protocol: 2,
	})

	// 启动阶段校验 Redis 是否可用，避免服务启动后才暴露依赖问题
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	fmt.Println("redis connect success")
	return client, nil
}
