package domain

import (
	"digital-wallet/internal/module"
	auth "digital-wallet/internal/module/user"
	webhook "digital-wallet/internal/module/webhook"
	"digital-wallet/platform"

	"digital-wallet/platform/logger"
)

type ModuleLayer struct {
	user    module.Auth
	tire    module.TireAccess
	webhook module.Event
}

func InitModule(persistence Persistence, log logger.Logger, plat platform.HTTPClient) ModuleLayer {
	return ModuleLayer{
		user: auth.Init(log.Named("auth-module"), persistence.user, persistence.catch),
		webhook: webhook.Init(log.Named("webhook-event"),
			persistence.user,
			persistence.event,
			persistence.account, persistence.tran, persistence.fee,
			persistence.bill, plat,
		),
	}
}
