package state

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type RetryParams struct {
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration
	MaxElapsedTime      time.Duration
}
type AuthDomains struct {
	User Domain
}

type Domain struct {
	Name string
	ID   string
}

func (a AuthDomains) Validate() error {
	return validation.Validate([]string{
		a.User.ID,
	}, validation.Each(validation.Required))
}
