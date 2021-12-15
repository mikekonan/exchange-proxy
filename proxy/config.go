package proxy

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Config struct {
	Port     string `help:"listen port"`
	Bindaddr string `help:"bindable address"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Port, is.Port),
		validation.Field(&c.Bindaddr, is.IPv4),
	)
}
