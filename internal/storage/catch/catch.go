package catch

import (
	"context"
	"digital-wallet/platform/logger"
	"fmt"
	"time"

	"digital-wallet/internal/constant/model/dto"
	st "digital-wallet/internal/storage"

	"github.com/go-redis/redis/v8"
	jwt "github.com/golang-jwt/jwt/v4"
)

type catch struct {
	logger      logger.Logger
	redis       *redis.Client
	jwtSecret   []byte
	redisClient *redis.Client
	sessionTTL  time.Duration
}

func Init(log logger.Logger, redis *redis.Client) st.Catch {
	return &catch{
		logger: log,
		redis:  redis,
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
