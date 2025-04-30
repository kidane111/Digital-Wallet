package storage

import (
	"context"
	"digital-wallet/internal/constant/model/dto"
	"time"

	"github.com/shopspring/decimal"
)

type Auth interface {
	UserRegister(ctx context.Context, param dto.RegisterRequest) (*dto.User, error)
	FindUserByPhone(ctx context.Context, phone string) (*dto.User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
}

type Catch interface {
	CreateSession(ctx context.Context, userID string, tier string) (*dto.Session, error)
}
type Account interface {
	UpdateBalance(ctx context.Context, userID string, amount decimal.Decimal) error
	UpdateBillStatus(ctx context.Context, billID string, status string) error
	GetBalance(ctx context.Context, userID string) (decimal.Decimal, error)
	// GetBillSummary(ctx context.Context, userID string) (map[string]interface{}, error)
	// GetTierLimits(ctx context.Context, tier string) (map[string]interface{}, error)
	GetTierConfig(ctx context.Context, tier string) (*dto.TierConfig, error)
	UpdateTierConfig(ctx context.Context, param dto.TierConfig) error
}
type EventRepository interface {
	Save(ctx context.Context, event dto.WebhookEvent) (*dto.WebhookEvent, error)
	Exists(ctx context.Context, eventID string) (bool, error)
	GetEventByEventID(ctx context.Context, eventID string) (*dto.WebhookEvent, error)
	UpdateStatus(
		ctx context.Context,
		eventID string,
		status dto.EventStatus,
		attempts int,
		lastAttemptAt time.Time,
		nextAttemptAt *time.Time,
		errorMessage *string,
	) error
}
type WalletRepository interface {
	Create(ctx context.Context, arg dto.WalletRequest) (*dto.WalletResponse, error)
	GetByID(ctx context.Context, id string) (*dto.WalletResponse, error)
	GetByUser(ctx context.Context, userID string, currency string) (*dto.WalletResponse, error)
	UpdateBalance(ctx context.Context, id string, balance decimal.Decimal) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tx dto.TransactionRequest) (*dto.TransactionResponse, error)
	GetByID(ctx context.Context, id string) (*dto.TransactionResponse, error)
	// GetByReference(ctx context.Context, reference string) (*dto.TransactionResponse, error)
	UpdateStatus(ctx context.Context, id string, status dto.TransactionStatus) error
	SumByType(ctx context.Context, walletID string, txType dto.TransactionType) (decimal.Decimal, error)
}

type FeeRepository interface {
	CalculateFee(ctx context.Context, arg dto.FeeCalculator) (*decimal.Decimal, error)
}

type Bill interface {
	RegisterBill(ctx context.Context, param dto.Payment) (*dto.Payment, error)
	GetBillByRefrenceID(ctx context.Context, refrenceID string) (*dto.Payment, error)
}
