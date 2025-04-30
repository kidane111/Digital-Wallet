package account

import (
	"context"
	"encoding/json"

	"digital-wallet/internal/constant/errors"
	"digital-wallet/internal/constant/model/db"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/model/persistencedb"
	st "digital-wallet/internal/storage"
	"digital-wallet/platform/logger"

	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type account struct {
	log            logger.Logger
	accountStorage persistencedb.PersistenceDB
}

func Init(
	log logger.Logger,
	accountStorage persistencedb.PersistenceDB,
) st.Account {
	return &account{
		log:            log,
		accountStorage: accountStorage,
	}
}

func (a *account) UpdateBillStatus(ctx context.Context, billID string, status string) error {
	err := a.accountStorage.UpdateBillStatus(ctx, db.UpdateBillStatusParams{
		ID:     billID,
		Status: status,
	})
	if err != nil {
		a.log.Error(ctx, "failed to update bill status", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update bill status")
	}
	return nil
}

func (a *account) GetBalance(ctx context.Context, userID string) (decimal.Decimal, error) {
	balance, err := a.accountStorage.GetBalanceByUserID(ctx, userID)
	if err != nil {
		a.log.Error(ctx, "failed to get balance", zap.Error(err))
		return decimal.Zero, errors.ErrUnableToGet.Wrap(err, "failed to get balance")
	}
	return balance, nil
}

func (a *account) GetTierConfig(ctx context.Context, tier string) (*dto.TierConfig, error) {
	dbConfig, err := a.accountStorage.GetTierConfig(ctx, tier)
	if err != nil {
		return nil, err
	}

	var features map[string]any
	if err := json.Unmarshal(dbConfig.Features.Bytes, &features); err != nil {
		return nil, err
	}

	return &dto.TierConfig{
		Tier:            dbConfig.Tier,
		Name:            dbConfig.Name,
		MonthlyFee:      dbConfig.MonthlyFee,
		MaxTransactions: int(dbConfig.MaxTransactions),
		MaxAmount:       dbConfig.MaxAmount,
		Features:        features,
		CreatedAt:       dbConfig.CreatedAt,
		UpdatedAt:       dbConfig.UpdatedAt,
	}, nil
}

func (a *account) UpdateTierConfig(ctx context.Context,
	param dto.TierConfig) error {
	featuresJson, err := json.Marshal(param.Features)
	if err != nil {
		return err
	}
	var featuresJSONB pgtype.JSONB
	if err := featuresJSONB.Set(featuresJson); err != nil {
		return err
	}
	err = a.accountStorage.UpdateTierConfig(ctx, db.UpdateTierConfigParams{
		Name:            param.Name,
		MonthlyFee:      param.MonthlyFee,
		MaxTransactions: int32(param.MaxTransactions),
		MaxAmount:       param.MaxAmount,
		Features:        featuresJSONB,
	})
	if err != nil {
		a.log.Error(ctx, "failed to update tier config", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update tier config")
	}
	return nil
}

func (a *account) UpdateBalance(ctx context.Context, userID string, amount decimal.Decimal) error {
	err := a.accountStorage.UpdateWalletBalance(ctx, db.UpdateWalletBalanceParams{
		ID:      userID,
		Balance: amount,
	})
	if err != nil {
		a.log.Error(ctx, "failed to update balance", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update balance")
	}
	return nil
}
