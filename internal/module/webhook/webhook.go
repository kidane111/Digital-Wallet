package webhook

import (
	"context"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/state"
	"digital-wallet/internal/module"
	"digital-wallet/internal/storage"
	platform "digital-wallet/platform"
	"digital-wallet/platform/logger"
	"digital-wallet/platform/retry"
	"net/http"

	"digital-wallet/platform/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"k8s.io/utils/pointer"
	// Removed unused import
)

type webhook struct {
	log             logger.Logger
	userStorage     storage.Auth
	event           storage.EventRepository
	account         storage.Account
	wallet          storage.WalletRepository
	transactionRepo storage.TransactionRepository
	fee             storage.FeeRepository
	bill            storage.Bill
	platform        platform.HTTPClient
}

func Init(log logger.Logger,
	userStorage storage.Auth,
	event storage.EventRepository,
	account storage.Account,
	tran storage.TransactionRepository,
	fee storage.FeeRepository, bill storage.Bill,
	http platform.HTTPClient) module.Event {
	return &webhook{
		log:             log,
		userStorage:     userStorage,
		event:           event,
		account:         account,
		transactionRepo: tran,
		fee:             fee,
		bill:            bill,
		platform:        http,
	}
}
func (w *webhook) HandleWebhook(ctx context.Context,
	param dto.Payload) (*dto.WebhookEvent, error) {
	// check if
	if exists, err := w.event.Exists(ctx, param.ID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("duplicate event")
	}
	event := dto.WebhookEvent{
		EventID: param.ID,
		Type:    dto.EventType(param.Type),
		Payload: param.Data,
	}
	data, err := w.event.Save(ctx, event)
	if err != nil {
		return nil, err
	}
	// start processing

	go w.processWithRetry(context.Background(), data.EventID)
	return data, nil
}

func (w *webhook) processWithRetry(ctx context.Context, eventID string) {
	event, err := w.event.GetEventByEventID(ctx, eventID)
	if err != nil {
		log.Printf("failed to get event: %v", err)
		return
	}

	retryParams := state.RetryParams{
		InitialInterval:     time.Second * 30,
		RandomizationFactor: 0.2,
		Multiplier:          2,
		MaxInterval:         time.Hour * 24,
		MaxElapsedTime:      time.Hour * 168,
	}

	// Push the job with retry logic
	retry.PushJob(ctx, w.log, retryParams, dto.Job[string]{
		Name: "process_wallet_transaction",
		Operation: func(ctx context.Context) (string, error) {
			err := w.processEvent(ctx, *event)
			if err != nil {
				return "", fmt.Errorf("failed to process event: %w", err)
			}
			return "", nil
		},
		OnSuccess: func(ctx context.Context, result string) error {
			// Update status to processed on success
			return w.event.UpdateStatus(
				ctx,
				event.ID,
				dto.EventStatusProcessed,
				int(retryParams.Multiplier), // Single successful attempt
				time.Now(),
				nil,
				nil,
			)
		},
		OnFailure: func(ctx context.Context, err error) error {
			// Update status to failed after all retries exhausted
			updateErr := w.event.UpdateStatus(
				ctx,
				event.ID,
				dto.EventStatusFailed,
				int(retryParams.Multiplier), // Single successful attempt
				time.Now(),
				nil,
				pointer.String(err.Error()),
			)
			if updateErr != nil {
				w.log.Error(ctx, "failed to update event status", zap.Error(updateErr))
			}

			// Optionally push to DLQ if needed
			// s.dlq.Store(ctx, *event)
			return nil
		},
	})

}

func (w *webhook) processEvent(ctx context.Context, event dto.WebhookEvent) error {
	// Validate the event first
	if err := validateEvent(event); err != nil {
		return fmt.Errorf("event validation failed: %w", err)
	}

	// Process based on event type
	switch event.Type {
	case dto.BillPayment:
		return w.processBillPayment(ctx, event)
	case dto.WalletTransaction:
		return w.processWalletTransaction(ctx, event)
	default:
		return errors.New("unsupported event type")
	}
}

func validateEvent(event dto.WebhookEvent) error {
	if event.ID == "" {
		return errors.New("event ID cannot be empty")
	}
	if event.CreatedAt.IsZero() {
		return errors.New("invalid event timestamp")
	}
	return nil
}
func (w *webhook) processBillPayment(ctx context.Context, event dto.WebhookEvent) error {

	var payment dto.Payment
	if err := json.Unmarshal(event.Payload, &payment); err != nil {
		return fmt.Errorf("failed to parse bill payment: %w", err)
	}

	// Validate payment data
	if payment.BillID == "" {
		return errors.New("bill ID cannot be empty")
	}
	if payment.Amount.GreaterThanOrEqual(decimal.Zero) {
		return errors.New("invalid payment amount")
	}

	// register the bill
	_, err := w.bill.RegisterBill(ctx, payment)
	if err != nil {
		return err
	}
	wallet, err := w.wallet.GetByUser(ctx, payment.UserID, "ETB")
	if err != nil {
		return err
	}
	// validate wallet balnce should be gretherthan amount
	if !wallet.Balance.GreaterThan(payment.Amount) {
		return errors.New("insufficent amount")

	}

	feerole, err := w.fee.CalculateFee(ctx, dto.FeeCalculator{
		Amount:          payment.Amount,
		Tire:            payment.UserTier,
		Time:            time.Now(),
		TransactionType: string(dto.TransactionBillPayment),
	})
	if err != nil {
		return fmt.Errorf("failed to get fee rules: %w", err)
	}
	if !wallet.Balance.GreaterThanOrEqual(*feerole) {
		return errors.New("insufficent amount")
	}

	// Check if payment already processed
	exists, err := w.bill.GetBillByRefrenceID(ctx, payment.ReferenceID)
	if err != nil {
		return fmt.Errorf("failed to check payment existence: %w", err)
	}
	if exists != nil {
		return fmt.Errorf("failed to check payment existence: %w", err)

	}

	_, err = w.transactionRepo.Create(ctx, dto.TransactionRequest{
		WalletID:    wallet.ID,
		Amount:      payment.Amount,
		Type:        payment.BillID,
		Description: "it a bill payment",
		Metadata: map[string]interface{}{
			"fee":          feerole,
			"user_tier":    payment.UserTier,
			"total_charge": feerole.Add(payment.Amount),
		}})
	if err != nil {
		return fmt.Errorf("failed to record payment: %w", err)
	}

	// Process payment (debit account)
	if err := w.wallet.UpdateBalance(ctx, wallet.ID, wallet.Balance.Sub(feerole.Add(payment.Amount))); err != nil {
		// Compensating action if debit fails
		_ = w.account.UpdateBillStatus(ctx, payment.ReferenceID, "failed")
		return fmt.Errorf("failed to debit account: %w", err)
	}

	// Update payment status to completed
	if err := w.account.UpdateBillStatus(ctx, payment.ReferenceID, "completed"); err != nil {
		w.log.Error(ctx, "failed to update payment status", zap.Error(err))
	}

	// Notify success
	return w.Notify(ctx, event)
}

func (w *webhook) processWalletTransaction(ctx context.Context, event dto.WebhookEvent) error {
	// Parse wallet transaction
	var transaction struct {
		WalletID        string          `json:"wallet_id"`
		Amount          decimal.Decimal `json:"amount"`
		Type            string          `json:"type"` // "credit" or "debit"
		Description     string          `json:"description"`
		Reference       string          `json:"reference"`
		TransferTo      string          `json:"transfer_to"`
		TransactionType string          `json:"transaction_type"`
		UserTier        string          `json:"user_tier"` // Added missing field
	}

	if err := json.Unmarshal(event.Payload, &transaction); err != nil {
		return fmt.Errorf("failed to parse wallet transaction: %w", err)
	}

	// Validate transaction
	switch {
	case transaction.WalletID == "":
		return errors.New("wallet ID cannot be empty")
	case transaction.Amount.LessThanOrEqual(decimal.NewFromInt(0)):
		return errors.New("invalid transaction amount")
	case transaction.Type != "credit" && transaction.Type != "debit":
		return errors.New("invalid transaction type")
	}

	var transferToWallet *dto.WalletResponse
	if transaction.TransactionType == string(dto.TransactionTransfer) {
		if transaction.TransferTo == "" {
			return errors.New("transfer_to wallet ID is required for transfer transactions")
		}

		var err error
		transferToWallet, err = w.wallet.GetByID(ctx, transaction.TransferTo)
		if err != nil {
			return fmt.Errorf("failed to get transfer_to wallet: %w", err)
		}
	}

	// Check wallet existence
	wallet, err := w.wallet.GetByID(ctx, transaction.WalletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	// Calculate fee
	feeRole, err := w.fee.CalculateFee(ctx, dto.FeeCalculator{
		Amount:          transaction.Amount,
		Tire:            transaction.UserTier,
		Time:            time.Now(),
		TransactionType: transaction.TransactionType, // Use actual transaction type
	})
	if err != nil {
		return fmt.Errorf("failed to calculate fee: %w", err)
	}

	var newBalance decimal.Decimal

	switch transaction.TransactionType {
	case string(dto.TransactionTransfer):
		if transferToWallet == nil {
			return errors.New("recipient wallet not found for transfer")
		}
		wallet1Update, wallet2Update, err := utils.MakeTransactionWithBalance(
			ctx,
			wallet.Balance,
			transferToWallet.Balance,
			transaction.Amount,
		)
		if err != nil {
			return fmt.Errorf("transfer failed: %w", err)
		}
		// Update both wallets
		if err := w.wallet.UpdateBalance(ctx, wallet.ID, wallet1Update); err != nil {
			return fmt.Errorf("failed to update sender wallet: %w", err)
		}
		if err := w.wallet.UpdateBalance(ctx, transferToWallet.ID, wallet2Update); err != nil {
			return fmt.Errorf("failed to update recipient wallet: %w", err)
		}
	case string(dto.TransactionDeposit):
		newBalance = wallet.Balance.Add(transaction.Amount).Sub(*feeRole)

	case string(dto.TransactionWithdrawal):
		newBalance = wallet.Balance.Sub(transaction.Amount).Sub(*feeRole)

	default:
		return fmt.Errorf("unsupported transaction type: %s", transaction.TransactionType)
	}

	// For non-transfer transactions, update the wallet balance
	if transaction.TransactionType != string(dto.TransactionTransfer) {
		if err := w.wallet.UpdateBalance(ctx, wallet.ID, newBalance); err != nil {
			return fmt.Errorf("failed to update wallet balance: %w", err)
		}
	}

	// Record transaction
	_, err = w.transactionRepo.Create(ctx, dto.TransactionRequest{
		WalletID:    wallet.ID,
		Amount:      transaction.Amount,
		Type:        transaction.Type,
		Description: transaction.Description,
		Metadata: map[string]interface{}{
			"fee":          feeRole,
			"user_tier":    transaction.UserTier,
			"total_charge": feeRole.Add(transaction.Amount),
			"transfer_to":  transferToWallet,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	// Update event status
	event.Status = "processed"
	return w.Notify(ctx, event)
}

func (w *webhook) Notify(ctx context.Context, event dto.WebhookEvent) error {

	// 2. Process based on event status
	switch event.Status {
	case string(dto.EventStatusProcessed):
		return w.handleSuccessfulEvent(ctx, event)
	case string(dto.EventStatusFailed):
		return w.handleFailedEvent(ctx, event)
	default:
		return nil
	}
}

func (w *webhook) handleSuccessfulEvent(ctx context.Context, event dto.WebhookEvent) error {
	// 1. Send to webhook callback URL if exists
	if event.CallbackUrl != "" {
		payload := map[string]interface{}{
			"event_id": event.ID,
			"status":   "success",
			"data":     event.Payload, // Assuming Payload contains the relevant data
		}

		retry.PushJob(ctx, w.log, state.RetryParams{
			InitialInterval:     time.Second * 30,
			RandomizationFactor: 0.2,
			Multiplier:          5,
			MaxInterval:         time.Hour * 24,
			MaxElapsedTime:      time.Hour * 168,
		}, dto.Job[string]{
			Name: "callback notification",
			Operation: func(ctx context.Context) (string, error) {
				_, err := w.platform.DoRequest(ctx, http.MethodPost, "content_type json", event.CallbackUrl,
					func(req *http.Request) {
						req.Header.Set("Content-Type", "application/json")
					}, payload, nil)
				if err != nil {
					w.log.Error(ctx, "unable to send request", zap.Error(err))
					return "", err
				}
				return "", nil
			},
			OnSuccess: func(ctx context.Context, result string) error {
				return nil
			},
			OnFailure: func(ctx context.Context, err error) error {
				w.log.Error(ctx, "unable to send sms", zap.Error(err))
				return nil
			},
		})
	}
	// even we can send the users to notify the satus uisng sms or email service in here

	return nil

}

func (w *webhook) handleFailedEvent(ctx context.Context, event dto.WebhookEvent) error {
	// 1. Send to webhook callback URL if exists
	if event.CallbackUrl != "" {
		payload := map[string]interface{}{
			"event_id": event.ID,
			"status":   "fail",
			"data":     event.Payload, // Assuming Payload contains the relevant data
		}

		retry.PushJob(ctx, w.log, state.RetryParams{
			InitialInterval:     time.Second * 30,
			RandomizationFactor: 0.2,
			Multiplier:          5,
			MaxInterval:         time.Hour * 24,
			MaxElapsedTime:      time.Hour * 168,
		}, dto.Job[string]{
			Name: "callback notification",
			Operation: func(ctx context.Context) (string, error) {
				_, err := w.platform.DoRequest(ctx, http.MethodPost, "content_type json", event.CallbackUrl,
					func(req *http.Request) {
						req.Header.Set("Content-Type", "application/json")
					}, payload, nil)
				if err != nil {
					w.log.Error(ctx, "unable to send request", zap.Error(err))
					return "", err
				}
				return "", nil
			},
			OnSuccess: func(ctx context.Context, result string) error {
				return nil
			},
			OnFailure: func(ctx context.Context, err error) error {
				w.log.Error(ctx, "unable to send sms", zap.Error(err))
				return nil
			},
		})
	}
	// even we can send the users to notify the satus uisng sms or email service in here

	return nil

}
