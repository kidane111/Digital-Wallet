package platform

import (
	"digital-wallet/internal/constant/state"
	"digital-wallet/platform"
	httpclient "digital-wallet/platform/callback"
	"digital-wallet/platform/logger"
)

func InitHTTPClient(httpConfig state.HTTPTransport, log logger.Logger) platform.HTTPClient {
	return httpclient.Init(httpConfig, log)
}
