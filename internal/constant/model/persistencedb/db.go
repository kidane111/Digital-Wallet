package persistencedb

import (
	"digital-wallet/internal/constant/model/db"
	"digital-wallet/platform/logger"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PersistenceDB struct {
	*db.Queries
	pool *pgxpool.Pool
	log  logger.Logger
}

func New(pool *pgxpool.Pool, log logger.Logger) PersistenceDB {
	return PersistenceDB{
		Queries: db.New(pool),
		pool:    pool,
		log:     log,
	}
}
