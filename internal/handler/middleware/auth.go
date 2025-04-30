package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"digital-wallet/internal/module"
	"digital-wallet/platform/logger"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type AuthMiddleware interface {
	Authenticate() gin.HandlerFunc
	TierAccessControl([]string) gin.HandlerFunc
	RateLimit() gin.HandlerFunc
	VarifaySignature() gin.HandlerFunc
}

type authMiddleware struct {
	logger                logger.Logger
	autz                  module.Catch
	authMiddlewareModules AuthMiddlewareModules
	redisClient           *redis.Client
	limiter               *rate.Limiter
}
type AuthMiddlewareModules struct {
	Auth module.Auth
}

func InitAuthMiddleware(
	logger logger.Logger,
	authMiddlewareModules AuthMiddlewareModules,
	redisClient *redis.Client, limiter int,
) AuthMiddleware {
	// map all the allowed tiers
	return &authMiddleware{
		logger:                logger,
		authMiddlewareModules: authMiddlewareModules,
		redisClient:           redisClient,
		limiter:               rate.NewLimiter(rate.Limit(limiter), limiter),
	}
}

func (a *authMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "authorization header required"})
			return
		}

		userID, tier, err := a.autz.ValidateSession(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "invalid session", "details": err.Error()})
			return
		}

		// Add user info to context
		c.Set("user_id", userID)
		c.Set("tier", tier)

		c.Next()
	}
}

func (a *authMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "unauthorized"})
			return
		}

		key := fmt.Sprintf("rate_limit:%s", userID.(string))
		ctx := c.Request.Context()

		// Use Redis for distributed rate limiting
		val, err := a.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"error": "internal server error"})
			return
		}

		if val >= 10 { // Allow 10 requests per window
			c.AbortWithStatusJSON(http.StatusTooManyRequests,
				gin.H{"error": "rate limit exceeded"})
			return
		}

		// Increment counter
		_, err = a.redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"error": "internal server error"})
			return
		}

		// Set expiry if this is the first request in the window
		if val == 0 {
			a.redisClient.Expire(ctx, key, time.Minute)
		}

		c.Next()
	}
}

func (a *authMiddleware) TierAccessControl(permission []string) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get tier from context (set by previous auth middleware)
		tier, exists := c.Get("tier")
		if !exists {
			a.logger.Warn(c.Request.Context(), "Tier not found in context")
			c.AbortWithStatusJSON(403, gin.H{
				"error":   "access_denied",
				"message": "Tier information not available",
			})
			return
		}

		tierStr, ok := tier.(string)
		if !ok {
			ctx := c.Request.Context()
			a.logger.Error(ctx, "Invalid tier type in context",
				zap.String("tier", "tier"), zap.Any("value", tier))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "server_error",
				"message": "Invalid tier information",
			})
			return
		}

		// Check if user's tier is allowed
		allowed := false
		for _, allowedTier := range permission {
			if tierStr == allowedTier {
				allowed = true
				break
			}
		}

		if !allowed {
			ctx := c.Request.Context()
			a.logger.Warn(ctx, "Access denied for tier",
				zap.String("tier", tierStr),
				zap.Strings("allowed_tiers", permission))
			c.AbortWithStatusJSON(403, gin.H{
				"error":          "access_denied",
				"message":        "Your tier does not have access to this resource",
				"required_tiers": permission,
			})
			return
		}

		c.Next()
	}
}

func (a *authMiddleware) VarifaySignature() gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("X-Signature")
		if signature == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest,
				 gin.H{"error": "missing signature"})
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				 gin.H{"error": "failed to read request body"})
			return
		}
		// Restore the body so it can be read again
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		secretKey := os.Getenv("SECRET_KEY")
		if secretKey == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"error": "server configuration error: missing secret key"})
			return
		}
		mac := hmac.New(sha256.New, []byte(secretKey))
		mac.Write(body)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}

		c.Next()
	}
}
