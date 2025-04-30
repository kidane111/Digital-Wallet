package domain

import (
	rest "digital-wallet/internal/handler"
	auth "digital-wallet/internal/handler/rest/authz"
	"digital-wallet/internal/handler/rest/tire"
	"digital-wallet/internal/handler/rest/webhook"

	"digital-wallet/platform/logger"
	"time"
)

type HandlerLayer struct {
	autzHandler rest.Auth
	tire        rest.Tire
	webhook     rest.Event
}

func InitHandler(module ModuleLayer, log logger.Logger, timeout time.Duration) HandlerLayer {
	return HandlerLayer{
		autzHandler: auth.Init(
			log.Named("auth-handler"),
			module.user,
			timeout,
		),
		tire: tire.Init(
			log.Named("tire-handler"),
			module.tire,
			timeout,
		),
		webhook: webhook.Init(log.Named("webhook-handler"),
			module.webhook, timeout),
	}
}
