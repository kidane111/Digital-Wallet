package state

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type HTTPTransport struct {
	MaxIdleConnsPerHost int
	MaxIdleConns        int
	MaxConnsPerHost     int
	Timeout             time.Duration
}

func (a HTTPTransport) Validate() error {
	return validation.Validate([]int{
		a.MaxConnsPerHost, a.MaxIdleConns, a.MaxIdleConnsPerHost, int(a.Timeout),
	}, validation.Each(validation.Required))
}
