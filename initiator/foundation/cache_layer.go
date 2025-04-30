package foundation

import (
	"digital-wallet/platform/logger"
	"fmt"

	"digital-wallet/internal/module/catch"
	st "digital-wallet/internal/storage"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

type CacheLayer struct {
	Redis *redis.Client
	catch st.Catch
}

type CacheOptions struct {
	Redis   *redis.Client
	account st.Account
}

func InitCacheLayer(cacheOptions CacheOptions, log logger.Logger) *CacheLayer {
	return &CacheLayer{
		Redis: cacheOptions.Redis,
		catch: catch.Init(
			log.Named("catch"),
			cacheOptions.Redis,
			[]byte(fmt.Sprintf("%d", viper.GetSizeInBytes("redis.jwt"))),
			viper.GetDuration("redis.expire_duration"),
			cacheOptions.account,
		),
	}
}
