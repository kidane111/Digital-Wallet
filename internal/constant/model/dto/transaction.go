package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
)

type TransactionRequest struct {
	WalletID    string                 `json:"wallet_id" validate:"required,uuid"`
	Amount      decimal.Decimal        `json:"amount" validate:"required,gt=0"`
	Type        string                 `json:"type" validate:"required,oneof=deposit withdrawal transfer bill_payment"`
	Description string                 `json:"description" validate:"required,max=255"`
	Reference   string                 `json:"reference" validate:"required,uuid"`
	Metadata    map[string]interface{} `json:"metadata"`
	Status      string
}

type TransactionType string

const (
	TransactionDeposit     TransactionType = "deposit"
	TransactionWithdrawal  TransactionType = "withdrawal"
	TransactionTransfer    TransactionType = "transfer"
	TransactionBillPayment TransactionType = "bill_payment"
)

type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "pending"
	TransactionCompleted TransactionStatus = "completed"
	TransactionFailed    TransactionStatus = "failed"
	TransactionReversed  TransactionStatus = "reversed"
)

type TransactionResponse struct {
	ID          string          `json:"id"`
	WalletID    string          `json:"wallet_id"`
	Amount      decimal.Decimal `json:"amount"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Reference   string          `json:"reference"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Ledgers struct {
	ID                   uuid.UUID       `json:"id,omitempty"`
	TransactionID        uuid.UUID       `json:"transaction_id,omitempty"`
	AccountID            string          `json:"account_id" validate:"required"`
	CreditDebitIndicator string          `json:"credit_debit_indicator" validate:"required"`
	Amount               decimal.Decimal `json:"amount" validate:"required"`
	CreatedAt            time.Time       `json:"created_at,omitempty"`
}
type MakeTransaction struct {
	Transaction *Transaction `json:"transaction" validate:"required"`
	Ledger      []Ledgers    `json:"ledger_entries" validate:"required"`
}

type Transaction struct {
	ID                     uuid.UUID         `json:"id"`
	Type                   TransactionType   `json:"type"`
	Status                 TransactionStatus `json:"status"`
	Note                   string            `json:"note"`
	PaymentReferenceNumber string            `json:"payment_reference_number"`
	ProductPaymentID       string            `json:"product_payment_id" validate:"required"`
	Detail                 pgtype.JSON       `json:"detail" validate:"required,json"`
	CreatedAt              time.Time         `json:"created_at"`
}
