package bill

import (
	"context"
	"digital-wallet/internal/constant/errors"
	"digital-wallet/internal/constant/errors/sqlcerr"
	"encoding/json"

	"digital-wallet/internal/constant/model/db"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/model/persistencedb"
	st "digital-wallet/internal/storage"
	"digital-wallet/platform/logger"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

type bill struct {
	log         logger.Logger
	billStorage persistencedb.PersistenceDB
}

func Init(
	log logger.Logger,
	billStorage persistencedb.PersistenceDB,
) st.Bill {
	return &bill{
		log:         log,
		billStorage: billStorage,
	}
}

// RegisterBill handles payment registration
func (s *bill) RegisterBill(ctx context.Context, param dto.Payment) (*dto.Payment, error) {
	existing, err := s.billStorage.RegisterBill(ctx, db.RegisterBillParams{
		UserID:        param.UserID,
		AccountNumber: param.Account,
		Amount:        param.Amount,
		Fee:           param.FeeAmount,
		Currency:      "ETB",
		Status:        param.Status,
		// DueDate:       param.dueDate,
		Metadata: pgtype.JSONB{
			Bytes:  []byte("{}"),
			Status: pgtype.Null,
		},
	})

	if param.Metadata != nil {
		metadataBytes, err := json.Marshal(param.Metadata)
		if err != nil {
			return nil, errors.ErrUnableToCreate.Wrap(err, "failed to marshal metadata")
		}
		existing.Metadata = pgtype.JSONB{
			Bytes:  metadataBytes,
			Status: pgtype.Present,
		}
	}
	if err != nil {
		return nil, err
	}

	return &dto.Payment{
		BillID:      existing.ID,
		Account:     existing.AccountNumber,
		Amount:      existing.Amount,
		FeeAmount:   existing.Amount,
		Currency:    existing.Currency,
		ReferenceID: existing.ID,
		Status:      existing.Status,
		ProcessedAt: existing.CreatedAt,
	}, nil
}

// GetBillByRefrenceID retrieves a payment by its reference ID
func (s *bill) GetBillByRefrenceID(ctx context.Context, referenceID string) (*dto.Payment, error) {
	payment, err := s.billStorage.GetBillByReferenceID(ctx, referenceID)
	if sqlcerr.Is(err, sqlcerr.ErrNoRows) {
		err = errors.ErrResourceNotFound.Wrap(err, "payment")
		s.log.Error(ctx, "payment not found",
			zap.Error(err),
			zap.Any("reference-id", referenceID))
		return nil, err
	}
	err = errors.ErrUnableToGet.Wrap(err, "error while getting billing")
	s.log.Error(ctx, "uable to get bill info usssing refrence id", zap.Error(err),
		zap.Any("reference-id", referenceID))

	return &dto.Payment{
		BillID:      payment.ID,
		Account:     payment.AccountNumber,
		Amount:      payment.Amount,
		FeeAmount:   payment.Amount,
		Currency:    payment.Currency,
		ReferenceID: payment.ID,
		Status:      payment.Status,
		ProcessedAt: payment.CreatedAt,
	}, nil
}
