package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

var ctx = context.Background()

func InitRedis() {

	rdb := redis.NewClient(&redis.Options{
		Addr: GetEnv("REDIS_HOST"),
		// Password: GetEnv("REDIS_PASSWORD"),
		// DB: 10,
	})

	err := rdb.Set(ctx, "Backend Started", time.Now().String(), 0).Err()
	if err != nil {
		panic("Redis connection failed")
	} else {
		fmt.Println("Connected to Redis!")
	}

	RedisClient = rdb
}
