package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type WalletRequest struct {
	UserID   string `json:"user_id" validate:"required,uuid"`
	Currency string `json:"currency" validate:"required,oneof=USD EUR GBP"`
}

type WalletResponse struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Balance   decimal.Decimal `json:"balance"`
	Currency  string          `json:"currency"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
