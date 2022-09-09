package helpers

import (
"github.com/go-redis/redis/v8"
"crypto/tls"
)

func InitRedis(){

	rdb := null

	if GetEnv("REDIS_HOST")!=""{
		rdb = redis.NewClient(&redis.Options{
			Addr:	  GetEnv("REDIS_HOST"),
			Password: GetEnv("REDIS_PASSWORD"),
			DB:		  GetEnv("REDIS_DB"),
		})
	}else if GetEnv("REDIS_DOMAIN"){
		rdb = redis.NewClient(&redis.Options{
			Password: GetEnv("REDIS_PASSWORD"),
			DB:		  GetEnv("REDIS_DB"),
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				ServerName: GetEnv("REDIS_DOMAIN"),
			},
		}) 
	} else {
		fmt.Println("Not using REDIS!")
	}

	return rdb
}