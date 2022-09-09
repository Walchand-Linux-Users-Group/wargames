package main

import (
	"github.com/Walchand-Linux-Users-Group/wargames/api/helpers"
)

func main() {

	helpers.InitEnv()

	mongoClient := helpers.InitDatabase()

	redisClient := helpers.InitRedis()

	initAPI(mongoClient, redisClient)
}
