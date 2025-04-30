package tire

import (
	"context"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/module"
	"digital-wallet/internal/storage"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisAdapter struct {
	client  *redis.Client
	catch   storage.Catch
	account storage.Account
}

func NewRedisAdapter(client *redis.Client, catch storage.Catch, account storage.Account) module.TireAccess {
	return &RedisAdapter{
		client:  client,
		catch:   catch,
		account: account,
	}
}

func (r *RedisAdapter) GetTierLimits(ctx context.Context, tier string) (*dto.TierConfig, error) {
	key := "tier_limits:" + tier
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var limits dto.TierConfig
	if err := json.Unmarshal([]byte(val), &limits); err != nil {
		return nil, err
	}
	// update in databse also
	//  data, err := r.account.CreateSession(ct)

	return &limits, nil
}

func (r *RedisAdapter) SetTierLimits(ctx context.Context, tier string, limits *dto.TierConfig) error {
	key := "tier_limits:" + tier
	jsonData, err := json.Marshal(limits)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, jsonData, 24*time.Hour).Err()
}

func (r *RedisAdapter) UpdateTierLimits(ctx context.Context,
	config dto.TierConfig) error {

	log.Printf("Updating tier limits for tier: %s", config.Tier)

	return r.account.UpdateTierConfig(ctx, config)

}
