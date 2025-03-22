// redis.go
package utils

import (
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

// function to initialize redis client to interact with redis
func InitRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return rdb
}

// function to load lua scripts as a redis script
func LoadScript(filename string) *redis.Script {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("error loading lua scripts %v", err)
	}
	return redis.NewScript(string(data))
}
