package transaction

import (
	"context"
	"digital-wallet/internal/constant/errors"
	"digital-wallet/internal/constant/errors/sqlcerr"
	"digital-wallet/internal/constant/model/db"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/model/persistencedb"
	st "digital-wallet/internal/storage"
	"digital-wallet/platform/logger"
	"encoding/json"

	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type transaction struct {
	log              logger.Logger
	transactionStore persistencedb.PersistenceDB
}

func Init(
	log logger.Logger,
	transactionStore persistencedb.PersistenceDB,
) st.TransactionRepository {
	return &transaction{
		log:              log,
		transactionStore: transactionStore,
	}
}

func (s *transaction) Create(ctx context.Context, tx dto.TransactionRequest) (*dto.TransactionResponse, error) {
	res, err := s.transactionStore.CreateTransaction(ctx, db.CreateTransactionParams{
		WalletID:        tx.WalletID,
		Amount:          tx.Amount,
		TransactionType: string(tx.Type),
		Status:          string(tx.Status),
		Description:     tx.Description,
		Metadata: func() pgtype.JSONB {
			metadataBytes, err := json.Marshal(tx.Metadata)
			if err != nil {
				s.log.Error(ctx, "failed to marshal metadata", zap.Error(err))
				return pgtype.JSONB{Status: pgtype.Null}
			}
			return pgtype.JSONB{Bytes: metadataBytes, Status: pgtype.Present}
		}(),
	})
	if err != nil {
		s.log.Error(ctx, "failed to create transaction", zap.Error(err))
		return nil, errors.ErrUnableToCreate.Wrap(err, "create transaction failed")
	}

	return &dto.TransactionResponse{
		ID:          res.ID,
		WalletID:    res.WalletID,
		Reference:   res.Reference.String(),
		Amount:      res.Amount,
		Type:        res.TransactionType,
		CreatedAt:   res.CreatedAt,
		Description: res.Description,
		Status:      res.Status,
		UpdatedAt:   res.UpdatedAt,
	}, nil
}

func (s *transaction) GetByID(ctx context.Context, id string) (*dto.TransactionResponse, error) {
	res, err := s.transactionStore.GetTransaction(ctx, id)
	if sqlcerr.Is(err, sqlcerr.ErrNoRows) {
		s.log.Error(ctx, "transaction not found", zap.Error(err), zap.String("id", id))
		return nil, errors.ErrResourceNotFound.Wrap(err, "transaction not found")
	}
	if err != nil {
		s.log.Error(ctx, "failed to get transaction by ID", zap.Error(err))
		return nil, errors.ErrUnableToGet.Wrap(err, "get transaction failed")
	}

	return &dto.TransactionResponse{
		ID:          res.ID,
		WalletID:    res.WalletID,
		Reference:   res.Reference.String(),
		Amount:      res.Amount,
		Type:        res.TransactionType,
		CreatedAt:   res.CreatedAt,
		Description: res.Description,
		Status:      res.Status,
		UpdatedAt:   res.UpdatedAt,
	}, nil
}

func (s *transaction) UpdateStatus(ctx context.Context, id string, status dto.TransactionStatus) error {
	err := s.transactionStore.UpdateTransactionStatus(ctx, db.UpdateTransactionStatusParams{
		ID:     id,
		Status: string(status),
	})
	if err != nil {
		s.log.Error(ctx, "failed to update transaction status", zap.Error(err), zap.String("id", id))
		return errors.ErrUnableToUpdate.Wrap(err, "update transaction status failed")
	}
	return nil
}

func (s *transaction) SumByType(ctx context.Context,
	walletID string, txType dto.TransactionType) (decimal.Decimal, error) {
	total, err := s.transactionStore.SumTransactionsByType(ctx,
		db.SumTransactionsByTypeParams{
			WalletID:        walletID,
			TransactionType: string(txType),
		})
	if err != nil {
		s.log.Error(ctx, "failed to sum transactions", zap.Error(err), zap.String("walletID", walletID))
		return decimal.Zero, errors.ErrUnableToGet.Wrap(err, "sum transaction failed")
	}
	decimalTotal, ok := total.(decimal.Decimal)
	if !ok {
		s.log.Error(ctx, "failed to assert total to decimal.Decimal", zap.String("walletID", walletID))
		return decimal.Zero, errors.ErrUnableToGet.New("type assertion to decimal.Decimal failed")
	}
	return decimalTotal, nil
}
