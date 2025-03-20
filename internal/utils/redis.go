package utils

import (
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return rdb
}

func LoadScript(filename string) *redis.Script {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("error loading lua scripts %v", err)
	}
	return redis.NewScript(string(data))
}
