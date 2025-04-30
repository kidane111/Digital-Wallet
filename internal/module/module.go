package module

import (
	"context"
	"digital-wallet/internal/constant/model/dto"
	"time"
)

type Auth interface {
	Login(ctx context.Context, param dto.Login) (*dto.LoginResponse, error)
	UserRegister(ctx context.Context, param dto.RegisterRequest) (*dto.User, error)
}

type Catch interface {
	CreateSession(ctx context.Context, userID string, tier string) (*dto.Session, error)
	ValidateSession(ctx context.Context, tokenString string) (string, string, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	PublishWriteOperation(ctx context.Context, op dto.WriteOperation) error
	SubscribeToWriteOperations(ctx context.Context) (<-chan dto.WriteOperation, error)
	AddToDeadLetterQueue(ctx context.Context, item dto.DeadLetterItem) error
	Start(ctx context.Context) error
	Stop() error
}

type TireAccess interface {
	// GetTierLimitsFromDB(ctx context.Context, tier string) (*dto.TierConfig, error)
	UpdateTierLimits(ctx context.Context, config dto.TierConfig) error
	GetTierLimits(ctx context.Context, tier string) (*dto.TierConfig, error)
	SetTierLimits(ctx context.Context, tier string, limits *dto.TierConfig) error
}
type Event interface {
	HandleWebhook(ctx context.Context, param dto.Payload) (*dto.WebhookEvent, error)
}
