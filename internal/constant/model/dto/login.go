package dto

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dongri/phonenumber"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/shopspring/decimal"
)

type Login struct {
	// Username of the user to login
	// @example	"john_doe"
	Username string `json:"username" `
	// Password of the user to login
	// @example	"password123"
	Password string `json:"password" `
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
	Tier      string    `json:"tier"`
}

func (l Login) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.Username, validation.Required, validation.Length(3, 50)),
		validation.Field(&l.Password, validation.Required, validation.Length(6, 100)),
	)
}

type User struct {
	ID string `json:"id"`
	//  Username of the user
	// @example	"john_doe"
	// @example	"john_doe"
	Username string `json:"username"`
	//  Email of the user
	Email string `json:"email"`
	//  FirstName of the user
	// @example	"John"
	FirstName string `json:"first_name"`
	//  LastName of the user
	LastName string `json:"last_name"`
	//  Phone of the user
	Phone string `json:"phone"`
	//  Password of the user
	// @example	"password123"
	Password string `json:"password"`
	//  Status of the user
	// @example	"active"
	Status string `json:"status"`
	//  Tier of the user
	// @example	"basic"
	Tier string `json:"tier"`
	//  CreatedAt of the user
	// @example	"2023-01-01T00:00:00Z"
	CreatedAt time.Time `json:"created_at"`
	//  UpdatedAt of the user
	// @example	"2023-01-01T00:00:00Z"
	UpdatedAt time.Time `json:"updated_at"`
	//  LastLogin of the user
	// @example	"2023-01-01T00:00:00Z"
	LastLogin time.Time `json:"last_login"`
}

type Session struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expired_at"`
}

// swagger:model
type RegisterRequest struct {
	// Name of the user
	// example: John Doe
	Name string `json:"name"`
	// Email address of the user
	// example: johndoe@example.com
	Email string `json:"email"`
	// Phone number of the user
	// example: +1234567890
	Phone string `json:"phone"`
	// Tier of the user account
	// example: Basic
	Tier string `json:"tier"`
	// Account creation timestamp
	// example: 2025-04-16T10:00:00Z
	Password  string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

func (u RegisterRequest) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.Required, validation.Length(3, 50)),
		validation.Field(&u.Email, validation.Required, validation.Length(6, 100)),
		validation.Field(&u.Phone, validation.Required.Error("phone is required"),
			validation.By(ValidatePhone)))
}
func ValidatePhone(phone any) error {
	str := phonenumber.Parse(fmt.Sprintf("%v", phone), "ET")
	if str == "" {
		return fmt.Errorf("invalid phone number")
	}

	return nil
}

type CacheItem struct {
	Key        string
	Value      interface{}
	Expiration time.Duration
}

type WriteOperation struct {
	ID        string
	Operation string
	Data      interface{}
	Retries   int
	CreatedAt time.Time
}

type DeadLetterItem struct {
	Operation  WriteOperation
	Error      string
	FailedAt   time.Time
	RetryCount int
}

type TierConfig struct {
	Tier            string          `json:"tier"`
	Name            string          `json:"name"`
	MonthlyFee      decimal.Decimal `json:"monthly_fee"`
	MaxTransactions int             `json:"max_transactions"`
	MaxAmount       decimal.Decimal `json:"max_amount"`
	Features        map[string]any  `json:"features"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type WebhookEvent struct {
	ID            string
	EventID       string
	Type          EventType
	CallbackUrl   string
	Status        string
	Payload       []byte
	Attempts      int
	LastAttemptAt time.Time
	NextAttemptAt *time.Time
	ErrorMessage  *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type EventType string

const (
	BillPayment       EventType = "bill_payment"
	WalletTransaction EventType = "wallet_transaction"
)

type EventStatus string

const (
	EventStatusPending   EventStatus = "pending"
	EventStatusProcessed EventStatus = "processed"
	EventStatusFailed    EventStatus = "failed"
)

type Payload struct {
	ID   string          `json:"event_id"`
	Type string          `json:"event_type"`
	Data json.RawMessage `json:"data"`
}

type Payment struct {
	BillID      string                 `json:"bill_id"`
	Account     string                 `json:"account"`
	Amount      decimal.Decimal        `json:"amount"`
	FeeAmount   decimal.Decimal        `json:"fee_amount"`
	Currency    string                 `json:"currency"`
	ReferenceID string                 `json:"reference_id"`
	Status      string                 `json:"status"`
	ProcessedAt time.Time              `json:"processed_at"`
	Metadata    map[string]interface{} `json:"metadata"`
	UserID      string                 `json:"user_id"`
	UserTier    string                 `json:"user_tire"`
}
