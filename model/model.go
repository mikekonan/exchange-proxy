package model

import (
	"time"
)

type Candle struct {
	Ts     time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
	Amount float64
}

func (c *Candle) Clone() *Candle {
	return &Candle{
		Ts:     c.Ts,
		Open:   c.Open,
		High:   c.High,
		Low:    c.Low,
		Close:  c.Close,
		Volume: c.Volume,
		Amount: c.Amount,
	}
}
