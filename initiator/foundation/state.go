package foundation

import (
	"context"
	"digital-wallet/internal/constant"
	"digital-wallet/internal/constant/state"
	"digital-wallet/platform/logger"
	"digital-wallet/platform/retry"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type State struct {
	RetryParams state.RetryParams
	AuthDomains state.AuthDomains
	HTTPConfig  state.HTTPTransport
}

func InitState(logger logger.Logger) State {

	authDomains := state.AuthDomains{
		User: state.Domain{
			ID:   viper.GetString("service.domain.user"),
			Name: string(constant.User),
		},
	}
	if err := authDomains.Validate(); err != nil {
		logger.Fatal(context.Background(), "invalid domain input",
			zap.Error(err),
		)
	}
	return State{
		RetryParams: retry.SetParams(state.RetryParams{
			InitialInterval:     viper.GetDuration("server.retry.initial_interval"),
			RandomizationFactor: viper.GetFloat64("server.retry.randomization_factor"),
			Multiplier:          viper.GetFloat64("server.retry.multiplier"),
			MaxInterval:         viper.GetDuration("server.retry.max_interval"),
			MaxElapsedTime:      viper.GetDuration("server.retry.max_elapsed_time"),
		}),
		AuthDomains: authDomains,
		HTTPConfig: state.HTTPTransport{
			MaxIdleConnsPerHost: viper.GetInt("rquest.max_conn_per_c"),
			MaxIdleConns:        int(viper.GetFloat64("server.retry.randomization_factor")),
			MaxConnsPerHost:     int(viper.GetDuration("server.retry.max_interval")),
			Timeout:             viper.GetDuration("request.timeout"),
		},
	}
}
