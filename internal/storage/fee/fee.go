package repository

import (
	"context"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/model/persistencedb"
	st "digital-wallet/internal/storage"
	"digital-wallet/platform/logger"

	"github.com/shopspring/decimal"
)

type fee struct {
	log logger.Logger
	db  persistencedb.PersistenceDB
}

func Init(
	log logger.Logger,
	db persistencedb.PersistenceDB,
) st.FeeRepository {
	return &fee{
		log: log,
		db:  db,
	}
}

func (r *fee) CalculateFee(ctx context.Context, arg dto.FeeCalculator) (*decimal.Decimal, error) {
	data, err := r.db.CalculateFee(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &data, nil

}
