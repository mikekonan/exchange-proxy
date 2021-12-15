package model

import (
	"bytes"
	"fmt"
	"strconv"
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

type Candles []*Candle

func f(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func (candles Candles) KucoinRespJSON() []byte {
	buff := bytes.NewBuffer(nil)
	buff.Write([]byte(`{"code":"200000","data":[`))
	for _, c := range candles {
		buff.Write([]byte(fmt.Sprintf(`["%d","%s","%s","%s","%s","%s","%s"],`, c.Ts.Unix(), f(c.Open), f(c.Close), f(c.High), f(c.Low), f(c.Volume), f(c.Amount))))
	}

	if len(candles) > 0 {
		buff.Truncate(buff.Len() - 1)
	}

	buff.Write([]byte(`]}`))

	return buff.Bytes()
}
