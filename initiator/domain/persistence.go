package domain

import (
	"digital-wallet/initiator/foundation"
	"digital-wallet/internal/constant/model/persistencedb"
	bill "digital-wallet/internal/storage/bill"
	event "digital-wallet/internal/storage/event"
	fee "digital-wallet/internal/storage/fee"
	trans "digital-wallet/internal/storage/transaction"

	wallet "digital-wallet/internal/storage/wallet"

	"digital-wallet/internal/module/catch"
	st "digital-wallet/internal/storage"
	"digital-wallet/internal/storage/account"
	user "digital-wallet/internal/storage/user"
	"fmt"

	"digital-wallet/platform/logger"

	"github.com/spf13/viper"
)

type Persistence struct {
	user    st.Auth
	catch   st.Catch
	account st.Account
	event   st.EventRepository
	fee     st.FeeRepository
	wallet  st.WalletRepository
	tran    st.TransactionRepository
	bill    st.Bill
}

func InitPersistence(db persistencedb.PersistenceDB, log logger.Logger,
	redis foundation.CacheLayer) Persistence {
	return Persistence{
		user: user.Init(log.Named("user"), db),
		catch: catch.Init(log.Named("catch"), redis.Redis,
			[]byte(fmt.Sprintf("%d", viper.GetSizeInBytes("redis.jwt"))),
			viper.GetDuration("redis.expire_duration"),
			account.Init(log.Named("account"), db)),
		event:  event.Init(log.Named("event-proecess"), db),
		fee:    fee.Init(log.Named("fee"), db),
		wallet: wallet.Init(log.Named("init-walle-persistance"), db),
		tran:   trans.Init(log.Named("transaction-init-peristance"), db),
		bill:   bill.Init(log.Named("bill-init-persistance"), db),
	}
}
