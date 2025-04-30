package foundation

import (
	"context"
	"digital-wallet/platform/logger"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func InitCache(url string, log logger.Logger) *redis.Client {
	opts, err := redis.ParseURL(url)
	if err != nil {
		log.Fatal(context.Background(), fmt.Sprintf("Failed to parse redis url: %v", err))
	}

	client := redis.NewClient(opts)

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		log.Fatal(context.Background(), fmt.Sprintf("Failed to ping redis: %v", err))
	}

	return client
}
