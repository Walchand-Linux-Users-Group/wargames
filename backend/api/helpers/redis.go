package helpers

import (
	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() {

	rdb := redis.NewClient(&redis.Options{
		Addr:     GetEnv("REDIS_HOST"),
		Password: GetEnv("REDIS_PASSWORD"),
		DB:       10,
	})

	RedisClient = rdb
}
