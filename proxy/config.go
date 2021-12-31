package proxy

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/valyala/fasthttp"
)

type Config struct {
	Port             string `help:"listen port"`
	Bindaddr         string `help:"bindable address"`
	ConcurrencyLimit int    `help:"server concurrency limit"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Port, is.Port),
		validation.Field(&c.Bindaddr, is.IPv4),
		validation.Field(&c.ConcurrencyLimit, validation.Min(fasthttp.DefaultConcurrency)),
	)
}
