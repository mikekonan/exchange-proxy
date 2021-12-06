package kucoin

import (
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Config struct {
	Port     string `help:"listen port"`
	Bindaddr string `help:"bindable address"`
	//Localaddr string `help:"local address (use it if you understand what you are doing)"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Port, is.Port),
		validation.Field(&c.Bindaddr, is.IPv4),
		//validation.Field(&c.Localaddr, validation.When(c.Localaddr != "", is.IPv4)),
	)
}
