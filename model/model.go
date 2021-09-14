package model

import (
	"bytes"
	"fmt"
)

type Candle struct {
	Exchange  string
	Pair      string
	Timeframe string
	Ts        int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Amount    float64
}

type Candles []*Candle

func (candles Candles) KucoinRespJSON() []byte {
	buff := bytes.NewBuffer(nil)
	buff.Write([]byte(`{"code":"200000","data":[`))
	for _, c := range candles {
		buff.Write([]byte(fmt.Sprintf(`["%d","%f","%f","%f","%f","%f","%f"],`, c.Ts, c.Open, c.Close, c.High, c.Low, c.Volume, c.Amount)))
	}

	buff.Truncate(buff.Len() - 1)
	buff.Write([]byte(`]}`))

	return buff.Bytes()
}
