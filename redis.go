package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	RDB *redis.Client
)

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Unable to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis")
}
