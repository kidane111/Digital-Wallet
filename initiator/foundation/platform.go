package foundation

import (
	"context"
	"digital-wallet/internal/constant/model/persistencedb"
	"digital-wallet/internal/constant/state"
	"digital-wallet/platform"
	plat "digital-wallet/platform/callback"

	"digital-wallet/platform/logger"
)

type PlatformLayer struct {
	http platform.HTTPClient
}

func InitPlatformLayer(
	_ context.Context,
	logger logger.Logger, _ persistencedb.PersistenceDB) PlatformLayer {
	return PlatformLayer{
		http: plat.Init(state.HTTPTransport{}, logger),
	}
}
