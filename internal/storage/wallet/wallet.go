package auth

import (
	"context"

	"digital-wallet/internal/constant/errors"
	"digital-wallet/internal/constant/errors/sqlcerr"
	"digital-wallet/internal/constant/model/db"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/model/persistencedb"
	st "digital-wallet/internal/storage"
	"digital-wallet/platform/logger"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type wallet struct {
	log logger.Logger
	db  persistencedb.PersistenceDB
}

func Init(
	log logger.Logger,
	db persistencedb.PersistenceDB,
) st.WalletRepository {
	return &wallet{
		log: log,
		db:  db,
	}
}

func (q *wallet) Create(ctx context.Context, arg dto.WalletRequest) (*dto.WalletResponse, error) {
	data, err := q.db.CreateWallet(ctx, db.CreateWalletParams{
		UserID:   arg.UserID,
		Currency: arg.Currency,
	})
	if err != nil {
		q.log.Error(ctx, "failed to update last login", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to update last login")
	}
	return &dto.WalletResponse{
		ID:        data.ID,
		UserID:    data.UserID,
		Balance:   data.Balance,
		Currency:  data.Currency,
		Status:    data.Status,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
	}, nil

}
func (q *wallet) GetByID(ctx context.Context, id string) (*dto.WalletResponse, error) {
	data, err := q.db.GetWallet(ctx, id)
	if err != nil {
		if sqlcerr.Is(err, sqlcerr.ErrNoRows) {
			q.log.Error(ctx, "user not found", zap.String("id", id))
			return nil, errors.ErrResourceNotFound.Wrap(err, "user not found")
		}
		err = errors.ErrUnableToGet.Wrap(err, "unable to check for program name")
		q.log.Error(ctx, "unable to check for program name", zap.Error(err))
		return nil, err
	}
	return &dto.WalletResponse{
		ID:        data.ID,
		UserID:    data.UserID,
		Balance:   data.Balance,
		Currency:  data.Currency,
		Status:    data.Status,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
	}, nil

}
func (q *wallet) GetByUser(ctx context.Context, userID string, currency string) (*dto.WalletResponse, error) {
	data, err := q.db.GetWalletByUser(ctx, db.GetWalletByUserParams{
		UserID:   userID,
		Currency: currency,
	})
	if err != nil {
		if sqlcerr.Is(err, sqlcerr.ErrNoRows) {
			q.log.Error(ctx, "user not found", zap.String("id", userID))
			return nil, errors.ErrResourceNotFound.Wrap(err, "user not found")
		}
		err = errors.ErrUnableToGet.Wrap(err, "unable to check for program name")
		q.log.Error(ctx, "unable to check for program name", zap.Error(err))
		return nil, err
	}
	return &dto.WalletResponse{
		ID:        data.ID,
		UserID:    data.UserID,
		Balance:   data.Balance,
		Currency:  data.Currency,
		Status:    data.Status,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
	}, nil
}
func (q *wallet) UpdateBalance(ctx context.Context, id string, balance decimal.Decimal) error {
	err := q.db.UpdateWalletBalance(ctx, db.UpdateWalletBalanceParams{
		ID:      id,
		Balance: balance,
	})
	if err != nil {
		err = errors.ErrUnableToGet.Wrap(err, "unable to check for program name")
		q.log.Error(ctx, "unable to check for program name", zap.Error(err))
		return err
	}
	// return &dto.WalletResponse{
	// 	ID:        data.ID,
	// 	UserID:    data.UserID,
	// 	Balance:   data.Balance,
	// 	Currency:  data.Currency,
	// 	Status:    data.Status,
	// 	CreatedAt: data.CreatedAt,
	// 	UpdatedAt: data.UpdatedAt,
	// }, nil
	return nil

}
