package database

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
)

var DBContext = context.Background()

func CreateClient(dbNo int) *redis.Client {
	redisDB := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DB_ADDRESS"),
		Password: os.Getenv("DB_PASSWORD"),
		DB:       dbNo,
	})
	return redisDB
}
