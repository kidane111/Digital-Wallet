package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type UserTier string

const (
	UserTierBasic      UserTier = "basic"
	UserTierPremium    UserTier = "premium"
	UserTierEnterprise UserTier = "enterprise"
)

type FeeCalculator struct {
	Amount          decimal.Decimal
	Tire            string
	Time            time.Time
	TransactionType string
}
