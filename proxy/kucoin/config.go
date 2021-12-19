package kucoin

import (
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Config struct {
	KucoinTopicsPerWs int    `help:"amount of topics per ws connection [10-280]"`
	KucoinApiURL      string `help:"kucoin api address"`
	//Localaddr string `help:"local address (use it if you understand what you are doing)"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.KucoinTopicsPerWs, validation.Min(10), validation.Max(280)),
		validation.Field(&c.KucoinApiURL, is.RequestURL),
		//validation.Field(&c.Localaddr, validation.When(c.Localaddr != "", is.IPv4)),
	)
}
