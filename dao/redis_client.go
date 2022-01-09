package dao

import (
	"context"
	"github.com/go-redis/redis/v8"
	"lyyops-cicd/config"
	"time"
)

var (
	RedisClient         *redis.Client
	sep                 = ":"
	defaultRedisTimeout = 60 * time.Second
)

func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Address,
		DB:       config.Config.Redis.Db,
		Password: config.Config.Redis.Password,
	})
	ctx, cancel := context.WithTimeout(context.Background(), defaultRedisTimeout) // 10秒超时
	defer cancel()
	return RedisClient.Ping(ctx).Err()
}
