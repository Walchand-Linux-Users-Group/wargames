package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func initEnv() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func main() {
	initEnv()

	_, err := GetMongoClient(getEnv("MONGO_URI"))

	if err != nil {
		log.Fatal(err)
	}

	err = clientInstance.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	initAPI()
}
