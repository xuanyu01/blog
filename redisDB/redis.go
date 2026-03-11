package redisDB

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var RedisClient *redis.Client

func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		panic(err)
		return err
	}

	fmt.Println("redis connect success")
	return nil
}
