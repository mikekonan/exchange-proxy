package model

import (
	"bytes"
	"fmt"
	"strconv"
)

type Candle struct {
	Exchange  string  `db:"exchange"`
	Pair      string  `db:"pair"`
	Timeframe string  `db:"timeframe"`
	Ts        int64   `db:"ts"`
	Open      float64 `db:"open"`
	High      float64 `db:"high"`
	Low       float64 `db:"low"`
	Close     float64 `db:"close"`
	Volume    float64 `db:"volume"`
	Amount    float64 `db:"amount"`
}

type Candles []Candle

func f(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func (candles Candles) KucoinRespJSON() []byte {
	buff := bytes.NewBuffer(nil)
	buff.Write([]byte(`{"code":"200000","data":[`))
	for _, c := range candles {
		buff.Write([]byte(fmt.Sprintf(`["%d","%s","%s","%s","%s","%s","%s"],`, c.Ts, f(c.Open), f(c.Close), f(c.High), f(c.Low), f(c.Volume), f(c.Amount))))
	}

	if len(candles) > 0 {
		buff.Truncate(buff.Len() - 1)
	}

	buff.Write([]byte(`]}`))

	return buff.Bytes()
}
