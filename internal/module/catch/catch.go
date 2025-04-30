package catch

import (
	"context"
	"digital-wallet/platform/logger"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/module"
	st "digital-wallet/internal/storage"

	"github.com/go-redis/redis/v8"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type catch struct {
	logger         logger.Logger
	jwtSecret      []byte
	redisClient    *redis.Client
	sessionTTL     time.Duration
	accountStorage st.Account
}

func Init(log logger.Logger, redis *redis.Client,
	secret []byte, sessionTTL time.Duration,
	accountStorage st.Account) module.Catch {
	return &catch{
		logger:         log,
		redisClient:    redis,
		jwtSecret:      secret,
		sessionTTL:     sessionTTL,
		accountStorage: accountStorage,
	}
}

func (s *catch) CreateSession(ctx context.Context, userID string, tier string) (*dto.Session, error) {
	// Create JWT token
	claims := jwt.MapClaims{
		"exp":     time.Now().Add(s.sessionTTL).Unix(),
		"user_id": userID,
		"tier":    tier,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	now := time.Now().Unix()
	sessionKey := fmt.Sprintf("session:%s", userID)
	_, err = s.redisClient.HSet(ctx, sessionKey,
		"token", tokenString,
		"tier", tier,
		"last_login", now,
	).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to store session in Redis: %w", err)
	}

	// Set TTL for the session
	_, err = s.redisClient.Expire(ctx, sessionKey, s.sessionTTL).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to set session TTL: %w", err)
	}

	return &dto.Session{
		Token:     tokenString,
		ExpiresAt: time.Now().Add(s.sessionTTL),
	}, nil
}

func (s *catch) ValidateSession(ctx context.Context, tokenString string) (string, string, error) {
	// Parse JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return "", "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", errors.New("invalid user_id in token")
	}

	tier, ok := claims["tier"].(string)
	if !ok {
		return "", "", errors.New("invalid tier in token")
	}

	// Verify session exists in Redis
	sessionKey := fmt.Sprintf("session:%s", userID)
	exists, err := s.redisClient.Exists(ctx, sessionKey).Result()
	if err != nil {
		return "", "", fmt.Errorf("failed to check session in Redis: %w", err)
	}
	if exists == 0 {
		return "", "", errors.New("session expired or not found")
	}

	return userID, tier, nil
}
func (s *catch) Get(ctx context.Context, key string) (string, error) {
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (s *catch) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.redisClient.Set(ctx, key, jsonData, expiration).Err()
}

func (s *catch) PublishWriteOperation(ctx context.Context, op dto.WriteOperation) error {
	jsonData, err := json.Marshal(op)
	if err != nil {
		return err
	}
	return s.redisClient.LPush(ctx, "write_operations", jsonData).Err()
}
func (s *catch) Start(ctx context.Context) error {
	opChan, err := s.SubscribeToWriteOperations(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case op := <-opChan:
			s.processOperation(ctx, op)
		}
	}
}

func (s *catch) processOperation(ctx context.Context, op dto.WriteOperation) {
	var err error

	// Retry with exponential backoff
	// Todo we can make this one the interval it should read from the config
	for i := 0; i < 3; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i*i) * time.Second) // Exponential backoff
		}

		switch op.Operation {
		case "update_balance":
			data := op.Data.(map[string]interface{})
			amount := decimal.NewFromFloat(data["amount"].(float64))
			err = s.accountStorage.UpdateBalance(ctx, data["user_id"].(string), amount)
		case "update_bill":
			data := op.Data.(map[string]interface{})
			err = s.accountStorage.UpdateBillStatus(ctx, data["bill_id"].(string), data["status"].(string))
		case "update_tier_config":
			config, ok := op.Data.(*dto.TierConfig)
			if !ok {
				s.logger.Error(ctx, "Invalid operation data type",
					zap.String("operation", op.Operation),
					zap.Any("data", op.Data))
				return
			}
			err = s.accountStorage.UpdateTierConfig(ctx, *config)
		}
		if err == nil {
			return // Success
		}
	}

	// All retries failed - add to dead letter queue
	if err := s.AddToDeadLetterQueue(ctx, dto.DeadLetterItem{
		Operation:  op,
		Error:      err.Error(),
		FailedAt:   time.Now(),
		RetryCount: op.Retries,
	}); err != nil {
		s.logger.Error(ctx, "Failed to add to dead letter queue",
			zap.Any("operation", op),
			zap.Error(err),
		)
	}
}

func (s *catch) SubscribeToWriteOperations(ctx context.Context) (<-chan dto.WriteOperation, error) {
	opChan := make(chan dto.WriteOperation)

	go func() {
		defer close(opChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				result, err := s.redisClient.BRPop(ctx, 0, "write_operations").Result()
				if err != nil {
					s.logger.Error(ctx, "Failed to read from operations queue", zap.Error(err))
					continue
				}
				var op dto.WriteOperation
				if err := json.Unmarshal([]byte(result[1]), &op); err != nil {
					s.logger.Error(ctx, "Failed to unmarshal operation", zap.Error(err))
					continue
				}

				opChan <- op
			}
		}
	}()

	return opChan, nil
}

func (s *catch) AddToDeadLetterQueue(ctx context.Context, item dto.DeadLetterItem) error {
	jsonData, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return s.redisClient.LPush(ctx, "dead_letter_queue", jsonData).Err()
}

func (s *catch) Stop() error {
	// Cleanup resources if needed
	return nil
}
